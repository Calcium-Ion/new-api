import React, { useEffect, useState } from 'react';
import {
  API,
  isMobile,
  shouldShowPrompt,
  showError,
  showInfo,
  showSuccess,
  showWarning,
  timestamp2string
} from '../helpers';

import { CHANNEL_OPTIONS, ITEMS_PER_PAGE } from '../constants';
import {
  getQuotaPerUnit,
  renderGroup,
  renderNumberWithPoint,
  renderQuota, renderQuotaWithPrompt
} from '../helpers/render';
import {
  Button, Divider,
  Dropdown,
  Form, Input,
  InputNumber, Modal,
  Popconfirm,
  Space,
  SplitButtonGroup,
  Switch,
  Table,
  Tag,
  Tooltip,
  Typography
} from '@douyinfe/semi-ui';
import EditChannel from '../pages/Channel/EditChannel';
import { IconList, IconTreeTriangleDown } from '@douyinfe/semi-icons';
import { loadChannelModels } from './utils.js';
import EditTagModal from '../pages/Channel/EditTagModal.js';
import TextNumberInput from './custom/TextNumberInput.js';

function renderTimestamp(timestamp) {
  return <>{timestamp2string(timestamp)}</>;
}

let type2label = undefined;

function renderType(type) {
  if (!type2label) {
    type2label = new Map();
    for (let i = 0; i < CHANNEL_OPTIONS.length; i++) {
      type2label[CHANNEL_OPTIONS[i].value] = CHANNEL_OPTIONS[i];
    }
    type2label[0] = { value: 0, text: '未知类型', color: 'grey' };
  }
  return (
    <Tag size="large" color={type2label[type]?.color}>
      {type2label[type]?.text}
    </Tag>
  );
}

function renderTagType(type) {
  return (
    <Tag
      color='light-blue'
      prefixIcon={<IconList />}
      size='large'
      shape='circle'
      type='light'
    >
      标签聚合
    </Tag>
  );
}

const ChannelsTable = () => {
  const columns = [
    // {
    //     title: '',
    //     dataIndex: 'checkbox',
    //     className: 'checkbox',
    // },
    {
      title: 'ID',
      dataIndex: 'id'
    },
    {
      title: '名称',
      dataIndex: 'name'
    },
    {
      title: '分组',
      dataIndex: 'group',
      render: (text, record, index) => {
        return (
          <div>
            <Space spacing={2}>
              {text?.split(',').map((item, index) => {
                return renderGroup(item);
              })}
            </Space>
          </div>
        );
      }
    },
    {
      title: '类型',
      dataIndex: 'type',
      render: (text, record, index) => {
        if (record.children === undefined) {
          return <>{renderType(text)}</>;
        } else {
          return <>{renderTagType(0)}</>;
        }
      }
    },
    {
      title: '状态',
      dataIndex: 'status',
      render: (text, record, index) => {
        if (text === 3) {
          if (record.other_info === '') {
            record.other_info = '{}';
          }
          let otherInfo = JSON.parse(record.other_info);
          let reason = otherInfo['status_reason'];
          let time = otherInfo['status_time'];
          return (
            <div>
              <Tooltip content={'原因：' + reason + '，时间：' + timestamp2string(time)}>
                {renderStatus(text)}
              </Tooltip>
            </div>
          );
        } else {
          return renderStatus(text);
        }
      }
    },
    {
      title: '响应时间',
      dataIndex: 'response_time',
      render: (text, record, index) => {
        return <div>{renderResponseTime(text)}</div>;
      }
    },
    {
      title: '已用/剩余',
      dataIndex: 'expired_time',
      render: (text, record, index) => {
        if (record.children === undefined) {
          return (
            <div>
              <Space spacing={1}>
                <Tooltip content={'已用额度'}>
                  <Tag color="white" type="ghost" size="large">
                    {renderQuota(record.used_quota)}
                  </Tag>
                </Tooltip>
                <Tooltip content={'剩余额度' + record.balance + '，点击更新'}>
                  <Tag
                    color="white"
                    type="ghost"
                    size="large"
                    onClick={() => {
                      updateChannelBalance(record);
                    }}
                  >
                    ${renderNumberWithPoint(record.balance)}
                  </Tag>
                </Tooltip>
              </Space>
            </div>
          );
        } else {
          return <Tooltip content={'已用额度'}>
            <Tag color="white" type="ghost" size="large">
              {renderQuota(record.used_quota)}
            </Tag>
          </Tooltip>;
        }
      }
    },
    {
      title: '优先级',
      dataIndex: 'priority',
      render: (text, record, index) => {
        if (record.children === undefined) {
          return (
            <div>
              <InputNumber
                style={{ width: 70 }}
                name="priority"
                onBlur={(e) => {
                  manageChannel(record.id, 'priority', record, e.target.value);
                }}
                keepFocus={true}
                innerButtons
                defaultValue={record.priority}
                min={-999}
              />
            </div>
          );
        } else {
          return <>
            <InputNumber
              style={{ width: 70 }}
              name="priority"
              keepFocus={true}
              onBlur={(e) => {
                Modal.warning({
                  title: '修改子渠道优先级',
                  content: '确定要修改所有子渠道优先级为 ' + e.target.value + ' 吗？',
                  onOk: () => {
                    if (e.target.value === '') {
                      return;
                    }
                    submitTagEdit('priority', {
                      tag: record.key,
                      priority: e.target.value
                    })
                  },
                })
              }}
              innerButtons
              defaultValue={record.priority}
              min={-999}
            />
          </>;
        }
      }
    },
    {
      title: '权重',
      dataIndex: 'weight',
      render: (text, record, index) => {
        if (record.children === undefined) {
          return (
            <div>
              <InputNumber
                style={{ width: 70 }}
                name="weight"
                onBlur={(e) => {
                  manageChannel(record.id, 'weight', record, e.target.value);
                }}
                keepFocus={true}
                innerButtons
                defaultValue={record.weight}
                min={0}
              />
            </div>
          );
        } else {
          return (
            <InputNumber
              style={{ width: 70 }}
              name="weight"
              keepFocus={true}
              onBlur={(e) => {
                Modal.warning({
                  title: '修改子渠道权重',
                  content: '确定要修改所有子渠道权重为 ' + e.target.value + ' 吗？',
                  onOk: () => {
                    if (e.target.value === '') {
                      return;
                    }
                    submitTagEdit('weight', {
                      tag: record.key,
                      weight: e.target.value
                    })
                  },
                })
              }}
              innerButtons
              defaultValue={record.weight}
              min={-999}
            />
          );
        }
      }
    },
    {
      title: '',
      dataIndex: 'operate',
      render: (text, record, index) => {
        if (record.children === undefined) {
          return (
            <div>
              <SplitButtonGroup
                style={{ marginRight: 1 }}
                aria-label="测试单个渠道操作项目组"
              >
                <Button
                  theme="light"
                  onClick={() => {
                    testChannel(record, '');
                  }}
                >
                  测试
                </Button>
                <Dropdown
                  trigger="click"
                  position="bottomRight"
                  menu={record.test_models}
                >
                  <Button
                    style={{ padding: '8px 4px' }}
                    type="primary"
                    icon={<IconTreeTriangleDown />}
                  ></Button>
                </Dropdown>
              </SplitButtonGroup>
              {/*<Button theme='light' type='primary' style={{marginRight: 1}} onClick={()=>testChannel(record)}>测试</Button>*/}
              <Popconfirm
                title="确定是否要删除此渠道？"
                content="此修改将不可逆"
                okType={'danger'}
                position={'left'}
                onConfirm={() => {
                  manageChannel(record.id, 'delete', record).then(() => {
                    removeRecord(record);
                  });
                }}
              >
                <Button theme="light" type="danger" style={{ marginRight: 1 }}>
                  删除
                </Button>
              </Popconfirm>
              {record.status === 1 ? (
                <Button
                  theme="light"
                  type="warning"
                  style={{ marginRight: 1 }}
                  onClick={async () => {
                    manageChannel(record.id, 'disable', record);
                  }}
                >
                  禁用
                </Button>
              ) : (
                <Button
                  theme="light"
                  type="secondary"
                  style={{ marginRight: 1 }}
                  onClick={async () => {
                    manageChannel(record.id, 'enable', record);
                  }}
                >
                  启用
                </Button>
              )}
              <Button
                theme="light"
                type="tertiary"
                style={{ marginRight: 1 }}
                onClick={() => {
                  setEditingChannel(record);
                  setShowEdit(true);
                }}
              >
                编辑
              </Button>
              <Popconfirm
                title="确定是否要复制此渠道？"
                content="复制渠道的所有信息"
                okType={'danger'}
                position={'left'}
                onConfirm={async () => {
                  copySelectedChannel(record);
                }}
              >
                <Button theme="light" type="primary" style={{ marginRight: 1 }}>
                  复制
                </Button>
              </Popconfirm>
            </div>
          );
        } else {
          return (
            <>
              <Button
                theme="light"
                type="secondary"
                style={{ marginRight: 1 }}
                onClick={async () => {
                  manageTag(record.key, 'enable');
                }}
              >
                启用全部
              </Button>
              <Button
                theme="light"
                type="warning"
                style={{ marginRight: 1 }}
                onClick={async () => {
                  manageTag(record.key, 'disable');
                }}
              >
                禁用全部
              </Button>
              <Button
                theme="light"
                type="tertiary"
                style={{ marginRight: 1 }}
                onClick={() => {
                  setShowEditTag(true);
                  setEditingTag(record.key);
                }}
              >
                编辑
              </Button>
            </>
          );
        }
      }
    }
  ];

  const [channels, setChannels] = useState([]);
  const [loading, setLoading] = useState(true);
  const [activePage, setActivePage] = useState(1);
  const [idSort, setIdSort] = useState(false);
  const [searchKeyword, setSearchKeyword] = useState('');
  const [searchGroup, setSearchGroup] = useState('');
  const [searchModel, setSearchModel] = useState('');
  const [searching, setSearching] = useState(false);
  const [updatingBalance, setUpdatingBalance] = useState(false);
  const [pageSize, setPageSize] = useState(ITEMS_PER_PAGE);
  const [showPrompt, setShowPrompt] = useState(
    shouldShowPrompt('channel-test')
  );
  const [channelCount, setChannelCount] = useState(pageSize);
  const [groupOptions, setGroupOptions] = useState([]);
  const [showEdit, setShowEdit] = useState(false);
  const [enableBatchDelete, setEnableBatchDelete] = useState(false);
  const [editingChannel, setEditingChannel] = useState({
    id: undefined
  });
  const [showEditTag, setShowEditTag] = useState(false);
  const [editingTag, setEditingTag] = useState('');
  const [selectedChannels, setSelectedChannels] = useState([]);
  const [showEditPriority, setShowEditPriority] = useState(false);
  const [enableTagMode, setEnableTagMode] = useState(false);


  const removeRecord = (record) => {
    let newDataSource = [...channels];
    if (record.id != null) {
      let idx = newDataSource.findIndex((data) => {
        if (data.children !== undefined) {
          for (let i = 0; i < data.children.length; i++) {
            if (data.children[i].id === record.id) {
              data.children.splice(i, 1);
              return false;
            }
          }
        } else {
          return data.id === record.id
        }
      });

      if (idx > -1) {
        newDataSource.splice(idx, 1);
        setChannels(newDataSource);
      }
    }
  };

  const setChannelFormat = (channels, enableTagMode) => {
    let channelDates = [];
    let channelTags = {};
    for (let i = 0; i < channels.length; i++) {
      channels[i].key = '' + channels[i].id;
      let test_models = [];
      channels[i].models.split(',').forEach((item, index) => {
        test_models.push({
          node: 'item',
          name: item,
          onClick: () => {
            testChannel(channels[i], item);
          }
        });
      });
      channels[i].test_models = test_models;
      if (!enableTagMode) {
        channelDates.push(channels[i]);
      } else {
        let tag = channels[i].tag;
        // find from channelTags
        let tagIndex = channelTags[tag];
        let tagChannelDates = undefined;
        if (tagIndex === undefined) {
          // not found, create a new tag
          channelTags[tag] = 1;
          tagChannelDates = {
            key: tag,
            id: tag,
            tag: tag,
            name: '标签：' + tag,
            group: '',
            used_quota: 0,
            response_time: 0,
            priority: -1,
            weight: -1,
          };
          tagChannelDates.children = [];
          channelDates.push(tagChannelDates);
        } else {
          // found, add to the tag
          tagChannelDates = channelDates.find((item) => item.key === tag);
        }
        if (tagChannelDates.priority === -1) {
          tagChannelDates.priority = channels[i].priority;
        } else {
          if (tagChannelDates.priority !== channels[i].priority) {
            tagChannelDates.priority = '';
          }
        }
        if (tagChannelDates.weight === -1) {
          tagChannelDates.weight = channels[i].weight;
        } else {
          if (tagChannelDates.weight !== channels[i].weight) {
            tagChannelDates.weight = '';
          }
        }

        if (tagChannelDates.group === '') {
          tagChannelDates.group = channels[i].group;
        } else {
          let channelGroupsStr = channels[i].group;
          channelGroupsStr.split(',').forEach((item, index) => {
            if (tagChannelDates.group.indexOf(item) === -1) {
              // join
              tagChannelDates.group += ',' + item;
            }
          });
        }

        tagChannelDates.children.push(channels[i]);
        if (channels[i].status === 1) {
          tagChannelDates.status = 1;
        }
        tagChannelDates.used_quota += channels[i].used_quota;
        tagChannelDates.response_time += channels[i].response_time;
        tagChannelDates.response_time = tagChannelDates.response_time / 2;
      }

    }
    // data.key = '' + data.id
    setChannels(channelDates);
    if (channelDates.length >= pageSize) {
      setChannelCount(channelDates.length + pageSize);
    } else {
      setChannelCount(channelDates.length);
    }
  };

  const loadChannels = async (startIdx, pageSize, idSort, enableTagMode) => {
    setLoading(true);
    const res = await API.get(
      `/api/channel/?p=${startIdx}&page_size=${pageSize}&id_sort=${idSort}&tag_mode=${enableTagMode}`
    );
    if (res === undefined) {
      return;
    }
    const { success, message, data } = res.data;
    if (success) {
      if (startIdx === 0) {
        setChannelFormat(data, enableTagMode);
      } else {
        let newChannels = [...channels];
        newChannels.splice(startIdx * pageSize, data.length, ...data);
        setChannelFormat(newChannels, enableTagMode);
      }
    } else {
      showError(message);
    }
    setLoading(false);
  };

  const copySelectedChannel = async (record) => {
    const channelToCopy = record
    channelToCopy.name += '_复制';
    channelToCopy.created_time = null;
    channelToCopy.balance = 0;
    channelToCopy.used_quota = 0;
    if (!channelToCopy) {
      showError('渠道未找到，请刷新页面后重试。');
      return;
    }
    try {
      const newChannel = { ...channelToCopy, id: undefined };
      const response = await API.post('/api/channel/', newChannel);
      if (response.data.success) {
        showSuccess('渠道复制成功');
        await refresh();
      } else {
        showError(response.data.message);
      }
    } catch (error) {
      showError('渠道复制失败: ' + error.message);
    }
  };

  const refresh = async () => {
    await loadChannels(activePage - 1, pageSize, idSort, enableTagMode);
  };

  useEffect(() => {
    // console.log('default effect')
    const localIdSort = localStorage.getItem('id-sort') === 'true';
    const localPageSize =
      parseInt(localStorage.getItem('page-size')) || ITEMS_PER_PAGE;
    setIdSort(localIdSort);
    setPageSize(localPageSize);
    loadChannels(0, localPageSize, localIdSort, enableTagMode)
      .then()
      .catch((reason) => {
        showError(reason);
      });
    fetchGroups().then();
    loadChannelModels().then();
  }, []);

  const manageChannel = async (id, action, record, value) => {
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
      if (action === 'delete') {
      } else {
        record.status = channel.status;
      }
      setChannels(newChannels);
    } else {
      showError(message);
    }
  };

  const manageTag = async (tag, action) => {
    console.log(tag, action);
    let res;
    switch (action) {
      case 'enable':
        res = await API.post('/api/channel/tag/enabled', {
          tag: tag
        });
        break;
      case 'disable':
        res = await API.post('/api/channel/tag/disabled', {
          tag: tag
        });
        break;
    }
    const { success, message } = res.data;
    if (success) {
      showSuccess('操作成功完成！');
      let newChannels = [...channels];
      for (let i = 0; i < newChannels.length; i++) {
        if (newChannels[i].tag === tag) {
          let status = action === 'enable' ? 1 : 2;
          newChannels[i]?.children?.forEach((channel) => {
            channel.status = status;
          });
          newChannels[i].status = status;
        }
      }
      setChannels(newChannels);
    } else {
      showError(message);
    }
  };

  const renderStatus = (status) => {
    switch (status) {
      case 1:
        return (
          <Tag size="large" color="green">
            已启用
          </Tag>
        );
      case 2:
        return (
          <Tag size="large" color="yellow">
            已禁用
          </Tag>
        );
      case 3:
        return (
          <Tag size="large" color="yellow">
            自动禁用
          </Tag>
        );
      default:
        return (
          <Tag size="large" color="grey">
            未知状态
          </Tag>
        );
    }
  };

  const renderResponseTime = (responseTime) => {
    let time = responseTime / 1000;
    time = time.toFixed(2) + ' 秒';
    if (responseTime === 0) {
      return (
        <Tag size="large" color="grey">
          未测试
        </Tag>
      );
    } else if (responseTime <= 1000) {
      return (
        <Tag size="large" color="green">
          {time}
        </Tag>
      );
    } else if (responseTime <= 3000) {
      return (
        <Tag size="large" color="lime">
          {time}
        </Tag>
      );
    } else if (responseTime <= 5000) {
      return (
        <Tag size="large" color="yellow">
          {time}
        </Tag>
      );
    } else {
      return (
        <Tag size="large" color="red">
          {time}
        </Tag>
      );
    }
  };

  const searchChannels = async (searchKeyword, searchGroup, searchModel) => {
    if (searchKeyword === '' && searchGroup === '' && searchModel === '') {
      await loadChannels(0, pageSize, idSort, enableTagMode);
      setActivePage(1);
      return;
    }
    setSearching(true);
    const res = await API.get(
      `/api/channel/search?keyword=${searchKeyword}&group=${searchGroup}&model=${searchModel}&id_sort=${idSort}&tag_mode=${enableTagMode}`
    );
    const { success, message, data } = res.data;
    if (success) {
      if (enableTagMode) {
        setChannelFormat(data, enableTagMode);
      } else {
        setChannels(data.map(channel => ({...channel, key: '' + channel.id})));
        setChannelCount(data.length);
      }
      setActivePage(1);
    } else {
      showError(message);
    }
    setSearching(false);
  };

  const testChannel = async (record, model) => {
    const res = await API.get(`/api/channel/test/${record.id}?model=${model}`);
    const { success, message, time } = res.data;
    if (success) {
      record.response_time = time * 1000;
      record.test_time = Date.now() / 1000;
      showInfo(`通道 ${record.name} 测试成功，耗时 ${time.toFixed(2)} 秒。`);
    } else {
      showError(message);
    }
  };

  const testAllChannels = async () => {
    const res = await API.get(`/api/channel/test`);
    const { success, message } = res.data;
    if (success) {
      showInfo('已成功开始测试所有通道，请刷新页面查看结果。');
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

  const updateChannelBalance = async (record) => {
    const res = await API.get(`/api/channel/update_balance/${record.id}/`);
    const { success, message, balance } = res.data;
    if (success) {
      record.balance = balance;
      record.balance_updated_time = Date.now() / 1000;
      showInfo(`通道 ${record.name} 余额更新成功！`);
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

  const batchDeleteChannels = async () => {
    if (selectedChannels.length === 0) {
      showError('请先选择要删除的通道！');
      return;
    }
    setLoading(true);
    let ids = [];
    selectedChannels.forEach((channel) => {
      ids.push(channel.id);
    });
    const res = await API.post(`/api/channel/batch`, { ids: ids });
    const { success, message, data } = res.data;
    if (success) {
      showSuccess(`已删除 ${data} 个通道！`);
      await refresh();
    } else {
      showError(message);
    }
    setLoading(false);
  };

  const fixChannelsAbilities = async () => {
    const res = await API.post(`/api/channel/fix`);
    const { success, message, data } = res.data;
    if (success) {
      showSuccess(`已修复 ${data} 个通道！`);
      await refresh();
    } else {
      showError(message);
    }
  };

  let pageData = channels.slice(
    (activePage - 1) * pageSize,
    activePage * pageSize
  );

  const handlePageChange = (page) => {
    setActivePage(page);
    if (page === Math.ceil(channels.length / pageSize) + 1) {
      // In this case we have to load more data and then append them.
      loadChannels(page - 1, pageSize, idSort, enableTagMode).then((r) => {
      });
    }
  };

  const handlePageSizeChange = async (size) => {
    localStorage.setItem('page-size', size + '');
    setPageSize(size);
    setActivePage(1);
    loadChannels(0, size, idSort, enableTagMode)
      .then()
      .catch((reason) => {
        showError(reason);
      });
  };

  const fetchGroups = async () => {
    try {
      let res = await API.get(`/api/group/`);
      // add 'all' option
      // res.data.data.unshift('all');
      if (res === undefined) {
        return;
      }
      setGroupOptions(
        res.data.data.map((group) => ({
          label: group,
          value: group
        }))
      );
    } catch (error) {
      showError(error.message);
    }
  };

  const submitTagEdit = async (type, data) => {
    switch (type) {
      case 'priority':
        if (data.priority === undefined || data.priority === '') {
          showInfo('优先级必须是整数！');
          return;
        }
        data.priority = parseInt(data.priority);
        break;
      case 'weight':
        if (data.weight === undefined || data.weight < 0 || data.weight === '') {
          showInfo('权重必须是非负整数！');
          return;
        }
        data.weight = parseInt(data.weight);
        break
    }

    try {
      const res = await API.put('/api/channel/tag', data);
      if (res?.data?.success) {
        showSuccess('更新成功！');
        await refresh();
      }
    } catch (error) {
      showError(error);
    }
  }

  const closeEdit = () => {
    setShowEdit(false);
  };

  const handleRow = (record, index) => {
    if (record.status !== 1) {
      return {
        style: {
          background: 'var(--semi-color-disabled-border)'
        }
      };
    } else {
      return {};
    }
  };

  return (
    <>
      <EditTagModal
        visible={showEditTag}
        tag={editingTag}
        handleClose={() => setShowEditTag(false)}
        refresh={refresh}
      />
      <EditChannel
        refresh={refresh}
        visible={showEdit}
        handleClose={closeEdit}
        editingChannel={editingChannel}
      />
      <Form
        onSubmit={() => {
          searchChannels(searchKeyword, searchGroup, searchModel);
        }}
        labelPosition="left"
      >
        <div style={{ display: 'flex' }}>
          <Space>
            <Form.Input
              field="search_keyword"
              label="搜索渠道关键词"
              placeholder="ID，名称和密钥 ..."
              value={searchKeyword}
              loading={searching}
              onChange={(v) => {
                setSearchKeyword(v.trim());
              }}
            />
            <Form.Input
              field="search_model"
              label="模型"
              placeholder="模型关键字"
              value={searchModel}
              loading={searching}
              onChange={(v) => {
                setSearchModel(v.trim());
              }}
            />
            <Form.Select
              field="group"
              label="分组"
              optionList={[{ label: '选择分组', value: null }, ...groupOptions]}
              initValue={null}
              onChange={(v) => {
                setSearchGroup(v);
                searchChannels(searchKeyword, v, searchModel);
              }}
            />
            <Button
              label="查询"
              type="primary"
              htmlType="submit"
              className="btn-margin-right"
              style={{ marginRight: 8 }}
            >
              查询
            </Button>
          </Space>
        </div>
      </Form>
      <Divider style={{ marginBottom: 15 }} />
      <div
        style={{
          display: isMobile() ? '' : 'flex',
          marginTop: isMobile() ? 0 : -45,
          zIndex: 999,
          pointerEvents: 'none'
        }}
      >
        <Space
          style={{ pointerEvents: 'auto', marginTop: isMobile() ? 0 : 45 }}
        >
          <Typography.Text strong>使用ID排序</Typography.Text>
          <Switch
            checked={idSort}
            label="使用ID排序"
            uncheckedText="关"
            aria-label="是否用ID排序"
            onChange={(v) => {
              localStorage.setItem('id-sort', v + '');
              setIdSort(v);
              loadChannels(0, pageSize, v, enableTagMode)
                .then()
                .catch((reason) => {
                  showError(reason);
                });
            }}
          ></Switch>
          <Button
            theme="light"
            type="primary"
            style={{ marginRight: 8 }}
            onClick={() => {
              setEditingChannel({
                id: undefined
              });
              setShowEdit(true);
            }}
          >
            添加渠道
          </Button>
          <Popconfirm
            title="确定？"
            okType={'warning'}
            onConfirm={testAllChannels}
            position={isMobile() ? 'top' : 'top'}
          >
            <Button theme="light" type="warning" style={{ marginRight: 8 }}>
              测试所有通道
            </Button>
          </Popconfirm>
          <Popconfirm
            title="确定？"
            okType={'secondary'}
            onConfirm={updateAllChannelsBalance}
          >
            <Button theme="light" type="secondary" style={{ marginRight: 8 }}>
              更新所有已启用通道余额
            </Button>
          </Popconfirm>
          <Popconfirm
            title="确定是否要删除禁用通道？"
            content="此修改将不可逆"
            okType={'danger'}
            onConfirm={deleteAllDisabledChannels}
          >
            <Button theme="light" type="danger" style={{ marginRight: 8 }}>
              删除禁用通道
            </Button>
          </Popconfirm>

          <Button
            theme="light"
            type="primary"
            style={{ marginRight: 8 }}
            onClick={refresh}
          >
            刷新
          </Button>
        </Space>
      </div>
      <div style={{ marginTop: 20 }}>
        <Space>
          <Typography.Text strong>开启批量删除</Typography.Text>
          <Switch
            label="开启批量删除"
            uncheckedText="关"
            aria-label="是否开启批量删除"
            onChange={(v) => {
              setEnableBatchDelete(v);
            }}
          ></Switch>
          <Popconfirm
            title="确定是否要删除所选通道？"
            content="此修改将不可逆"
            okType={'danger'}
            onConfirm={batchDeleteChannels}
            disabled={!enableBatchDelete}
            position={'top'}
          >
            <Button
              disabled={!enableBatchDelete}
              theme="light"
              type="danger"
              style={{ marginRight: 8 }}
            >
              删除所选通道
            </Button>
          </Popconfirm>
          <Popconfirm
            title="确定是否要修复数据库一致性？"
            content="进行该操作时，可能导致渠道访问错误，请仅在数据库出现问题时使用"
            okType={'warning'}
            onConfirm={fixChannelsAbilities}
            position={'top'}
          >
            <Button theme="light" type="secondary" style={{ marginRight: 8 }}>
              修复数据库一致性
            </Button>
          </Popconfirm>
        </Space>
      </div>
      <div style={{ marginTop: 20 }}>
      <Space>
          <Typography.Text strong>标签聚合模式</Typography.Text>
          <Switch
            checked={enableTagMode}
            label="标签聚合模式"
            uncheckedText="关"
            aria-label="是否启用标签聚合"
            onChange={(v) => {
              setEnableTagMode(v);
              // 切换模式时重新加载数据
              loadChannels(0, pageSize, idSort, v);
            }}
          />
        </Space>
      </div>


      <Table
        className={'channel-table'}
        style={{ marginTop: 15 }}
        columns={columns}
        dataSource={pageData}
        pagination={{
          currentPage: activePage,
          pageSize: pageSize,
          total: channelCount,
          pageSizeOpts: [10, 20, 50, 100],
          showSizeChanger: true,
          formatPageText: (page) => '',
          onPageSizeChange: (size) => {
            handlePageSizeChange(size).then();
          },
          onPageChange: handlePageChange
        }}
        loading={loading}
        onRow={handleRow}
        rowSelection={
          enableBatchDelete
            ? {
              onChange: (selectedRowKeys, selectedRows) => {
                // console.log(`selectedRowKeys: ${selectedRowKeys}`, 'selectedRows: ', selectedRows);
                setSelectedChannels(selectedRows);
              }
            }
            : null
        }
      />
    </>
  );
};

export default ChannelsTable;
