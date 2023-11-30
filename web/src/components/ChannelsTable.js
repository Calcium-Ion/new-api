import React, { useEffect, useState } from 'react';
import { Button, Form, Input, Label, Message, Pagination, Popup, Table } from 'semantic-ui-react';
import { Link } from 'react-router-dom';
import { API, setPromptShown, shouldShowPrompt, showError, showInfo, showSuccess, timestamp2string } from '../helpers';

import { CHANNEL_OPTIONS, ITEMS_PER_PAGE } from '../constants';
import {renderGroup, renderNumber, renderQuota} from '../helpers/render';

function renderTimestamp(timestamp) {
  return (
    <>
      {timestamp2string(timestamp)}
    </>
  );
}

let type2label = undefined;

function renderType(type) {
  if (!type2label) {
    type2label = new Map;
    for (let i = 0; i < CHANNEL_OPTIONS.length; i++) {
      type2label[CHANNEL_OPTIONS[i].value] = CHANNEL_OPTIONS[i];
    }
    type2label[0] = { value: 0, text: '未知类型', color: 'grey' };
  }
  return <Label basic color={type2label[type]?.color}>{type2label[type]?.text}</Label>;
}

function renderBalance(type, balance) {
  switch (type) {
    case 1: // OpenAI
      return <span>${balance.toFixed(2)}</span>;
    case 4: // CloseAI
      return <span>¥{balance.toFixed(2)}</span>;
    case 8: // 自定义
      return <span>${balance.toFixed(2)}</span>;
    case 5: // OpenAI-SB
      return <span>¥{(balance / 10000).toFixed(2)}</span>;
    case 10: // AI Proxy
      return <span>{renderNumber(balance)}</span>;
    case 12: // API2GPT
      return <span>¥{balance.toFixed(2)}</span>;
    case 13: // AIGC2D
      return <span>{renderNumber(balance)}</span>;
    default:
      return <span>不支持</span>;
  }
}

const ChannelsTable = () => {
  const [channels, setChannels] = useState([]);
  const [loading, setLoading] = useState(true);
  const [activePage, setActivePage] = useState(1);
  const [searchKeyword, setSearchKeyword] = useState('');
  const [searching, setSearching] = useState(false);
  const [updatingBalance, setUpdatingBalance] = useState(false);
  const [pageSize, setPageSize] = useState(ITEMS_PER_PAGE);
  const [showPrompt, setShowPrompt] = useState(shouldShowPrompt("channel-test"));

  const loadChannels = async (startIdx) => {
    const res = await API.get(`/api/channel/?p=${startIdx}&page_size=${pageSize}`);
    const { success, message, data } = res.data;
    if (success) {
      if (startIdx === 0) {
        setChannels(data);
      } else {
        let newChannels = [...channels];
        newChannels.splice(startIdx * pageSize, data.length, ...data);
        setChannels(newChannels);
      }
    } else {
      showError(message);
    }
    setLoading(false);
  };

  const onPaginationChange = (e, { activePage }) => {
    (async () => {
      if (activePage === Math.ceil(channels.length / pageSize) + 1) {
        // In this case we have to load more data and then append them.
        await loadChannels(activePage - 1, pageSize);
      }
      setActivePage(activePage);
    })();
  };

  const setItemsPerPage = (e) => {
    console.log(e.target.value);
    //parseInt(e.target.value);
    setPageSize(parseInt(e.target.value));
    loadChannels(0);
  }

  const refresh = async () => {
    setLoading(true);
    await loadChannels(activePage - 1);
  };

  useEffect(() => {
    loadChannels(0)
      .then()
      .catch((reason) => {
        showError(reason);
      });
  }, []);

  const manageChannel = async (id, action, idx, value) => {
    let data = { id };
    let res;
    switch (action) {
      case 'delete':
        res = await API.delete(`/api/channel/${id}/`);
        break;
      case 'enable':
        data.status = 1;
        res = await API.put('/api/channel/', data);
        break;
      case 'disable':
        data.status = 2;
        res = await API.put('/api/channel/', data);
        break;
      case 'priority':
        if (value === '') {
          return;
        }
        data.priority = parseInt(value);
        res = await API.put('/api/channel/', data);
        break;
      case 'weight':
        if (value === '') {
          return;
        }
        data.weight = parseInt(value);
        if (data.weight < 0) {
          data.weight = 0;
        }
        res = await API.put('/api/channel/', data);
        break;
    }
    const { success, message } = res.data;
    if (success) {
      showSuccess('操作成功完成！');
      let channel = res.data.data;
      let newChannels = [...channels];
      let realIdx = (activePage - 1) * pageSize + idx;
      if (action === 'delete') {
        newChannels[realIdx].deleted = true;
      } else {
        newChannels[realIdx].status = channel.status;
      }
      setChannels(newChannels);
    } else {
      showError(message);
    }
  };

  const renderStatus = (status) => {
    switch (status) {
      case 1:
        return <Label basic color='green'>已启用</Label>;
      case 2:
        return (
          <Popup
            trigger={<Label basic color='red'>
              已禁用
            </Label>}
            content='本渠道被手动禁用'
            basic
          />
        );
      case 3:
        return (
          <Popup
            trigger={<Label basic color='yellow'>
              已禁用
            </Label>}
            content='本渠道被程序自动禁用'
            basic
          />
        );
      default:
        return (
          <Label basic color='grey'>
            未知状态
          </Label>
        );
    }
  };

  const renderResponseTime = (responseTime) => {
    let time = responseTime / 1000;
    time = time.toFixed(2) + ' 秒';
    if (responseTime === 0) {
      return <Label basic color='grey'>未测试</Label>;
    } else if (responseTime <= 1000) {
      return <Label basic color='green'>{time}</Label>;
    } else if (responseTime <= 3000) {
      return <Label basic color='olive'>{time}</Label>;
    } else if (responseTime <= 5000) {
      return <Label basic color='yellow'>{time}</Label>;
    } else {
      return <Label basic color='red'>{time}</Label>;
    }
  };

  const searchChannels = async () => {
    if (searchKeyword === '') {
      // if keyword is blank, load files instead.
      await loadChannels(0);
      setActivePage(1);
      return;
    }
    setSearching(true);
    const res = await API.get(`/api/channel/search?keyword=${searchKeyword}`);
    const { success, message, data } = res.data;
    if (success) {
      setChannels(data);
      setActivePage(1);
    } else {
      showError(message);
    }
    setSearching(false);
  };

  const testChannel = async (id, name, idx) => {
    const res = await API.get(`/api/channel/test/${id}/`);
    const { success, message, time } = res.data;
    if (success) {
      let newChannels = [...channels];
      let realIdx = (activePage - 1) * pageSize + idx;
      newChannels[realIdx].response_time = time * 1000;
      newChannels[realIdx].test_time = Date.now() / 1000;
      setChannels(newChannels);
      showInfo(`通道 ${name} 测试成功，耗时 ${time.toFixed(2)} 秒。`);
    } else {
      showError(message);
    }
  };

  const testAllChannels = async () => {
    const res = await API.get(`/api/channel/test`);
    const { success, message } = res.data;
    if (success) {
      showInfo('已成功开始测试所有已启用通道，请刷新页面查看结果。');
    } else {
      showError(message);
    }
  };

  const deleteAllDisabledChannels = async () => {
    const res = await API.delete(`/api/channel/disabled`);
    const { success, message, data } = res.data;
    if (success) {
      showSuccess(`已删除所有禁用渠道，共计 ${data} 个`);
      await refresh();
    } else {
      showError(message);
    }
  };

  const updateChannelBalance = async (id, name, idx) => {
    const res = await API.get(`/api/channel/update_balance/${id}/`);
    const { success, message, balance } = res.data;
    if (success) {
      let newChannels = [...channels];
      let realIdx = (activePage - 1) * pageSize + idx;
      newChannels[realIdx].balance = balance;
      newChannels[realIdx].balance_updated_time = Date.now() / 1000;
      setChannels(newChannels);
      showInfo(`通道 ${name} 余额更新成功！`);
    } else {
      showError(message);
    }
  };

  const updateAllChannelsBalance = async () => {
    setUpdatingBalance(true);
    const res = await API.get(`/api/channel/update_balance`);
    const { success, message } = res.data;
    if (success) {
      showInfo('已更新完毕所有已启用通道余额！');
    } else {
      showError(message);
    }
    setUpdatingBalance(false);
  };

  const handleKeywordChange = async (e, { value }) => {
    setSearchKeyword(value.trim());
  };

  const sortChannel = (key) => {
    if (channels.length === 0) return;
    setLoading(true);
    let sortedChannels = [...channels];
    if (typeof sortedChannels[0][key] === 'string') {
      sortedChannels.sort((a, b) => {
        return ('' + a[key]).localeCompare(b[key]);
      });
    } else {
      sortedChannels.sort((a, b) => {
        if (a[key] === b[key]) return 0;
        if (a[key] > b[key]) return -1;
        if (a[key] < b[key]) return 1;
      });
    }
    if (sortedChannels[0].id === channels[0].id) {
      sortedChannels.reverse();
    }
    setChannels(sortedChannels);
    setLoading(false);
  };

  return (
    <>
      <Form onSubmit={searchChannels}>
        <Form.Input
          icon='search'
          fluid
          iconPosition='left'
          placeholder='搜索渠道的 ID，名称和密钥 ...'
          value={searchKeyword}
          loading={searching}
          onChange={handleKeywordChange}
        />
      </Form>
      {
        showPrompt && (
          <Message onDismiss={() => {
            setShowPrompt(false);
            setPromptShown("channel-test");
          }}>
            当前版本测试是通过按照 OpenAI API 格式使用 gpt-3.5-turbo
            模型进行非流式请求实现的，因此测试报错并不一定代表通道不可用，该功能后续会修复。

            另外，OpenAI 渠道已经不再支持通过 key 获取余额，因此余额显示为 0。对于支持的渠道类型，请点击余额进行刷新。
          </Message>
        )
      }
      <Table basic compact size='small'>
        <Table.Header>
          <Table.Row>
            <Table.HeaderCell
              style={{ cursor: 'pointer' }}
              onClick={() => {
                sortChannel('id');
              }}
            >
              ID
            </Table.HeaderCell>
            <Table.HeaderCell
              style={{ cursor: 'pointer' }}
              onClick={() => {
                sortChannel('name');
              }}
            >
              名称
            </Table.HeaderCell>
            <Table.HeaderCell
              style={{ cursor: 'pointer' }}
              onClick={() => {
                sortChannel('group');
              }}
              width={1}
            >
              分组
            </Table.HeaderCell>
            <Table.HeaderCell
              style={{ cursor: 'pointer' }}
              onClick={() => {
                sortChannel('type');
              }}
              width={2}
            >
              类型
            </Table.HeaderCell>
            <Table.HeaderCell
              style={{ cursor: 'pointer' }}
              onClick={() => {
                sortChannel('status');
              }}
              width={2}
            >
              状态
            </Table.HeaderCell>
            <Table.HeaderCell
              style={{ cursor: 'pointer' }}
              onClick={() => {
                sortChannel('response_time');
              }}
            >
              响应时间
            </Table.HeaderCell>
            <Table.HeaderCell
                style={{ cursor: 'pointer' }}
                onClick={() => {
                  sortChannel('used_quota');
                }}
                width={1}
            >
              已使用
            </Table.HeaderCell>
            <Table.HeaderCell
              style={{ cursor: 'pointer' }}
              onClick={() => {
                sortChannel('balance');
              }}
            >
              余额
            </Table.HeaderCell>
            <Table.HeaderCell
              style={{ cursor: 'pointer' }}
              onClick={() => {
                sortChannel('priority');
              }}
            >
              优先级
            </Table.HeaderCell>
            <Table.HeaderCell>操作</Table.HeaderCell>
          </Table.Row>
        </Table.Header>

        <Table.Body>
          {channels
            .slice(
              (activePage - 1) * pageSize,
              activePage * pageSize
            )
            .map((channel, idx) => {
              if (channel.deleted) return <></>;
              return (
                <Table.Row key={channel.id}>
                  <Table.Cell>{channel.id}</Table.Cell>
                  <Table.Cell>{channel.name ? channel.name : '无'}</Table.Cell>
                  <Table.Cell>{renderGroup(channel.group)}</Table.Cell>
                  <Table.Cell>{renderType(channel.type)}</Table.Cell>
                  <Table.Cell>{renderStatus(channel.status)}</Table.Cell>
                  <Table.Cell>
                    <Popup
                      content={channel.test_time ? renderTimestamp(channel.test_time) : '未测试'}
                      key={channel.id}
                      trigger={renderResponseTime(channel.response_time)}
                      basic
                    />
                  </Table.Cell>
                  <Table.Cell>{renderQuota(channel.used_quota)}</Table.Cell>
                  <Table.Cell>
                    <Popup
                      trigger={<span onClick={() => {
                        updateChannelBalance(channel.id, channel.name, idx);
                      }} style={{ cursor: 'pointer' }}>
                      {renderBalance(channel.type, channel.balance)}
                    </span>}
                      content='点击更新'
                      basic
                    />
                  </Table.Cell>
                  <Table.Cell>
                    <Popup
                      trigger={<Input type='number' defaultValue={channel.priority} onBlur={(event) => {
                        manageChannel(
                          channel.id,
                          'priority',
                          idx,
                          event.target.value
                        );
                      }}>
                        <input style={{ maxWidth: '60px' }} />
                      </Input>}
                      content='渠道选择优先级，越高越优先'
                      basic
                    />
                  </Table.Cell>
                  <Table.Cell>
                    <div>
                      <Button
                        size={'small'}
                        positive
                        onClick={() => {
                          testChannel(channel.id, channel.name, idx);
                        }}
                      >
                        测试
                      </Button>
                      <Popup
                        trigger={
                          <Button size='small' negative>
                            删除
                          </Button>
                        }
                        on='click'
                        flowing
                        hoverable
                      >
                        <Button
                          negative
                          onClick={() => {
                            manageChannel(channel.id, 'delete', idx);
                          }}
                        >
                          删除渠道 {channel.name}
                        </Button>
                      </Popup>
                      <Button
                        size={'small'}
                        onClick={() => {
                          manageChannel(
                            channel.id,
                            channel.status === 1 ? 'disable' : 'enable',
                            idx
                          );
                        }}
                      >
                        {channel.status === 1 ? '禁用' : '启用'}
                      </Button>
                      <Button
                        size={'small'}
                        as={Link}
                        to={'/channel/edit/' + channel.id}
                      >
                        编辑
                      </Button>
                    </div>
                  </Table.Cell>
                </Table.Row>
              );
            })}
        </Table.Body>

        <Table.Footer>
          <Table.Row>
            <Table.HeaderCell colSpan='10'>
              <Button size='small' as={Link} to='/channel/add' loading={loading}>
                添加新的渠道
              </Button>
              <Button size='small' loading={loading} onClick={testAllChannels}>
                测试所有已启用通道
              </Button>
              <Button size='small' onClick={updateAllChannelsBalance}
                      loading={loading || updatingBalance}>更新所有已启用通道余额</Button>

              <div style={{ float: 'right' }}>
                <div className="ui labeled input" style={{marginRight: '10px'}}>
                  <div className="ui label">每页数量</div>
                  <Input type="number" style={{width: '70px'}} defaultValue={ITEMS_PER_PAGE} onBlur={setItemsPerPage}></Input>
                </div>
                <Pagination
                    activePage={activePage}
                    onPageChange={onPaginationChange}
                    size='small'
                    siblingRange={1}
                    totalPages={
                        Math.ceil(channels.length / pageSize) +
                        (channels.length % pageSize === 0 ? 1 : 0)
                    }
                />
              </div>
              <Popup
                trigger={
                  <Button size='small' loading={loading}>
                    删除禁用渠道
                  </Button>
                }
                on='click'
                flowing
                hoverable
              >
                <Button size='small' loading={loading} negative onClick={deleteAllDisabledChannels}>
                  确认删除
                </Button>
              </Popup>
              <Button size='small' onClick={refresh} loading={loading}>刷新</Button>

            </Table.HeaderCell>
          </Table.Row>
        </Table.Footer>
      </Table>
    </>
  );
};

export default ChannelsTable;
