import React, { useEffect, useState } from 'react';
import { Button, Form, Header, Label, Pagination, Segment, Select, Table,Modal } from 'semantic-ui-react';
import { API, isAdmin, showError, timestamp2string } from '../helpers';

import { ITEMS_PER_PAGE } from '../constants';
import { renderQuota } from '../helpers/render';
import {Link} from "react-router-dom";

function renderTimestamp(timestamp) {
  return (
    <>
      {timestamp2string(timestamp)}
    </>
  );
}

const MODE_OPTIONS = [
  { key: 'all', text: '全部用户', value: 'all' },
  { key: 'self', text: '当前用户', value: 'self' }
];

const LOG_OPTIONS = [
  { key: '0', text: '全部', value: 0 },
  // { key: '1', text: '绘图', value: 1 },
  // { key: '2', text: '放大', value: 2 },
  // { key: '3', text: '变换', value: 3 },
  // { key: '4', text: '图生文', value: 4 },
  // { key: '5', text: '图片混合', value: 5 }
];

function renderType(type) {
  switch (type) {
    case 'IMAGINE':
      return <Label basic color='blue'> 绘图 </Label>;
    case 'UPSCALE':
      return <Label basic color='orange'> 放大 </Label>;
    case 'VARIATION':
      return <Label basic color='purple'> 变换 </Label>;
    case 'DESCRIBE':
      return <Label basic color='yellow'> 图生文 </Label>;
    case 'BLEAND':
      return <Label basic color='olive'> 图混合 </Label>;
    default:
      return <Label basic color='black'> 未知 </Label>;
  }
}

function renderCode(type) {
  switch (type) {
    case 1:
      return <Label basic color='green'> 已提交 </Label>;
    case 21:
      return <Label basic color='olive'> 排队中 </Label>;
    case 22:
      return <Label basic color='orange'> 重复提交 </Label>;
    default:
      return <Label basic color='black'> 未知 </Label>;
  }
}

function renderStatus(type) {
  switch (type) {
    case 'SUCCESS':
      return <Label basic color='green'> 成功 </Label>;
    case 'NOT_START':
      return <Label basic color='black'> 未启动 </Label>;
    case 'SUBMITTED':
      return <Label basic color='yellow'> 队列中 </Label>;
    case 'IN_PROGRESS':
      return <Label basic color='blue'> 执行中 </Label>;
    case 'FAILURE':
      return <Label basic color='red'> 失败 </Label>;
    default:
      return <Label basic color='black'> 未知 </Label>;
  }
}

const LogsTable = () => {
  const [logs, setLogs] = useState([

  ]);
  const [loading, setLoading] = useState(true);
  const [activePage, setActivePage] = useState(1);
  const [searchKeyword, setSearchKeyword] = useState('');
  const [searching, setSearching] = useState(false);
  const [logType, setLogType] = useState(0);
  const isAdminUser = isAdmin();


  let now = new Date();
  const [inputs, setInputs] = useState({
    username: '',
    token_name: '',
    model_name: '',
    start_timestamp: timestamp2string(0),
    end_timestamp: timestamp2string(now.getTime() / 1000 + 3600)
  });
  const { username, token_name, model_name, start_timestamp, end_timestamp } = inputs;

  const [stat, setStat] = useState({
    quota: 0,
    token: 0
  });

  const [modalContent, setModalContent] = useState('');
  const [showModal, setShowModal] = useState(false);

  const showFullContent = (content) => {
    setModalContent(content);
    setShowModal(true);
  };


  const loadLogs = async (startIdx) => {
    let url = '';
    let localStartTimestamp = Date.parse(start_timestamp) / 1000;
    let localEndTimestamp = Date.parse(end_timestamp) / 1000;
    if (isAdminUser) {
      url = `/api/mj/?p=${startIdx}&username=${username}&token_name=${token_name}&start_timestamp=${localStartTimestamp}&end_timestamp=${localEndTimestamp}`;
    } else {
      url = `/api/mj/self/?p=${startIdx}&token_name=${token_name}&start_timestamp=${localStartTimestamp}&end_timestamp=${localEndTimestamp}`;
    }
    const res = await API.get(url);
    const { success, message, data } = res.data;
    if (success) {
      if (startIdx === 0) {
        setLogs(data);
      } else {
        let newLogs = [...logs];
        newLogs.splice(startIdx * ITEMS_PER_PAGE, data.length, ...data);
        setLogs(newLogs);
      }
    } else {
      showError(message);
    }
    setLoading(false);
  };

  const onPaginationChange = (e, { activePage }) => {
    (async () => {
      if (activePage === Math.ceil(logs.length / ITEMS_PER_PAGE) + 1) {
        // In this case we have to load more data and then append them.
        await loadLogs(activePage - 1);
      }
      setActivePage(activePage);
    })();
  };

  const refresh = async () => {
    setLoading(true);
    setActivePage(1)
    await loadLogs(0);
    // if (isAdminUser) {
    //   getLogStat().then();
    // } else {
    //   getLogSelfStat().then();
    // }
  };

  useEffect(() => {
    refresh().then();
  }, [logType]);

  const searchLogs = async () => {
    if (searchKeyword === '') {
      // if keyword is blank, load files instead.
      await loadLogs(0);
      setActivePage(1);
      return;
    }
    setSearching(true);
    const res = await API.get(`/api/log/self/search?keyword=${searchKeyword}`);
    const { success, message, data } = res.data;
    if (success) {
      setLogs(data);
      setActivePage(1);
    } else {
      showError(message);
    }
    setSearching(false);
  };

  const handleKeywordChange = async (e, { value }) => {
    setSearchKeyword(value.trim());
  };

  const sortLog = (key) => {
    if (logs.length === 0) return;
    setLoading(true);
    let sortedLogs = [...logs];
    if (typeof sortedLogs[0][key] === 'string'){
      sortedLogs.sort((a, b) => {
        return ('' + a[key]).localeCompare(b[key]);
      });
    } else {
      sortedLogs.sort((a, b) => {
        if (a[key] === b[key]) return 0;
        if (a[key] > b[key]) return -1;
        if (a[key] < b[key]) return 1;
      });
    }
    if (sortedLogs[0].id === logs[0].id) {
      sortedLogs.reverse();
    }
    setLogs(sortedLogs);
    setLoading(false);
  };

  return (
    <>
      <Segment>
        <Table basic compact size='small'>
          <Table.Header>
            <Table.Row>
              <Table.HeaderCell
                style={{ cursor: 'pointer' }}
                onClick={() => {
                  sortLog('submit_time');
                }}
                width={2}
              >
                提交时间
              </Table.HeaderCell>
              <Table.HeaderCell
                  style={{ cursor: 'pointer' }}
                  onClick={() => {
                    sortLog('action');
                  }}
                  width={1}
              >
                类型
              </Table.HeaderCell>
              <Table.HeaderCell
                  style={{ cursor: 'pointer' }}
                  onClick={() => {
                    sortLog('mj_id');
                  }}
                  width={2}
              >
                任务ID
              </Table.HeaderCell>
              <Table.HeaderCell
                style={{ cursor: 'pointer' }}
                onClick={() => {
                  sortLog('code');
                }}
                width={1}
              >
                提交结果
              </Table.HeaderCell>
              <Table.HeaderCell
                  style={{ cursor: 'pointer' }}
                  onClick={() => {
                    sortLog('status');
                  }}
                  width={1}
              >
                任务状态
              </Table.HeaderCell>
              <Table.HeaderCell
                style={{ cursor: 'pointer' }}
                onClick={() => {
                  sortLog('progress');
                }}
                width={1}
              >
                进度
              </Table.HeaderCell>
              <Table.HeaderCell
                style={{ cursor: 'pointer' }}
                onClick={() => {
                  sortLog('image_url');
                }}
                width={1}
              >
                结果图片
              </Table.HeaderCell>
              <Table.HeaderCell
                style={{ cursor: 'pointer' }}
                onClick={() => {
                  sortLog('prompt');
                }}
                width={3}
              >
                Prompt
              </Table.HeaderCell>
              <Table.HeaderCell
                  style={{ cursor: 'pointer' }}
                  onClick={() => {
                    sortLog('prompt_en');
                  }}
                  width={3}
              >
                PromptEn
              </Table.HeaderCell>
              <Table.HeaderCell
                  style={{ cursor: 'pointer' }}
                  onClick={() => {
                    sortLog('fail_reason');
                  }}
                  width={1}
              >
                失败原因
              </Table.HeaderCell>
            </Table.Row>
          </Table.Header>

          <Table.Body>
            {logs
              .slice(
                (activePage - 1) * ITEMS_PER_PAGE,
                activePage * ITEMS_PER_PAGE
              )
              .map((log, idx) => {
                if (log.deleted) return <></>;
                return (
                  <Table.Row key={log.created_at}>
                    <Table.Cell>{renderTimestamp(log.submit_time/1000)}</Table.Cell>
                    {/*{*/}
                    {/*  isAdminUser && (*/}
                    {/*    <Table.Cell>{log.username ? <Label>{log.username}</Label> : ''}</Table.Cell>*/}
                    {/*  )*/}
                    {/*}*/}
                    <Table.Cell>{renderType(log.action)}</Table.Cell>
                    <Table.Cell>{log.mj_id}</Table.Cell>
                    <Table.Cell>{renderCode(log.code)}</Table.Cell>
                    <Table.Cell>{renderStatus(log.status)}</Table.Cell>
                    <Table.Cell>{log.progress ? <Label basic>{log.progress}</Label> : ''}</Table.Cell>
                    <Table.Cell>
                      {
                        log.image_url ? (
                            // <Link to={log.image_url} target='_blank'>点击查看</Link>
                            <a href={log.image_url} target='_blank'>点击查看</a>
                        ) : '暂未生成图片'
                      }
                    </Table.Cell>
                    <Table.Cell>
                      {log.prompt.length > 10
                          ? <div>
                            {log.prompt.slice(0, 10)}
                            <a onClick={() => showFullContent(log.prompt)}>查看全部</a>
                          </div>
                          : log.prompt
                      }
                    </Table.Cell>
                    <Table.Cell>
                      {log.prompt_en.length > 10
                          ? <div>
                            {log.prompt_en.slice(0, 10)}
                            <a onClick={() => showFullContent(log.prompt_en)}>查看全部</a>
                          </div>
                          : log.prompt_en
                      }
                    </Table.Cell>
                    <Table.Cell>
                      {log.fail_reason && log.fail_reason.length > 10
                          ? <div>
                            {log.fail_reason.slice(0, 10)}
                            <a onClick={() => showFullContent(log.fail_reason)}>查看全部</a>
                          </div>
                          : log.fail_reason || '无'
                      }
                    </Table.Cell>
                  </Table.Row>
                );
              })}
          </Table.Body>

          <Table.Footer>
            <Table.Row>
              <Table.HeaderCell colSpan={'9'}>
                <Select
                  placeholder='选择明细分类'
                  options={LOG_OPTIONS}
                  style={{ marginRight: '8px' }}
                  name='logType'
                  value={logType}
                  onChange={(e, { name, value }) => {
                    setLogType(value);
                  }}
                />
                <Button size='small' onClick={refresh} loading={loading}>刷新</Button>
                <Pagination
                  floated='right'
                  activePage={activePage}
                  onPageChange={onPaginationChange}
                  size='small'
                  siblingRange={1}
                  totalPages={
                    Math.ceil(logs.length / ITEMS_PER_PAGE) +
                    (logs.length % ITEMS_PER_PAGE === 0 ? 1 : 0)
                  }
                />
              </Table.HeaderCell>
            </Table.Row>
          </Table.Footer>
        </Table>
      </Segment>
      {/*Modal component goes here*/}
      <Modal open={showModal} onClose={() => setShowModal(false)} centered>
        <Modal.Content>
          <pre style={{whiteSpace: "pre-wrap"}}>{modalContent}</pre>
        </Modal.Content>
      </Modal>
    </>
  );
};

export default LogsTable;
