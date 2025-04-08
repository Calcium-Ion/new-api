import React, { useContext, useEffect, useState } from 'react';
import { useTranslation } from 'react-i18next';
import {
  API,
  copy,
  getTodayStartTimestamp,
  isAdmin,
  showError,
  showSuccess,
  timestamp2string,
} from '../helpers';

import {
  Avatar,
  Button,
  Descriptions,
  Form,
  Layout,
  Modal,
  Popover,
  Select,
  Space,
  Spin,
  Table,
  Tag,
  Tooltip,
  Checkbox,
} from '@douyinfe/semi-ui';
import { ITEMS_PER_PAGE } from '../constants';
import {
  renderAudioModelPrice,
  renderClaudeLogContent,
  renderClaudeModelPrice,
  renderClaudeModelPriceSimple,
  renderGroup,
  renderLogContent,
  renderModelPrice,
  renderModelPriceSimple,
  renderNumber,
  renderQuota,
  stringToColor,
} from '../helpers/render';
import Paragraph from '@douyinfe/semi-ui/lib/es/typography/paragraph';
import { getLogOther } from '../helpers/other.js';
import { StyleContext } from '../context/Style/index.js';
import { IconInherit, IconRefresh, IconSetting } from '@douyinfe/semi-icons';

const { Header } = Layout;

function renderTimestamp(timestamp) {
  return <>{timestamp2string(timestamp)}</>;
}

const MODE_OPTIONS = [
  { key: 'all', text: 'all', value: 'all' },
  { key: 'self', text: 'current user', value: 'self' },
];

const colors = [
  'amber',
  'blue',
  'cyan',
  'green',
  'grey',
  'indigo',
  'light-blue',
  'lime',
  'orange',
  'pink',
  'purple',
  'red',
  'teal',
  'violet',
  'yellow',
];

const LogsTable = () => {
  const { t } = useTranslation();

  function renderType(type) {
    switch (type) {
      case 1:
        return (
          <Tag color='cyan' size='large'>
            {t('充值')}
          </Tag>
        );
      case 2:
        return (
          <Tag color='lime' size='large'>
            {t('消费')}
          </Tag>
        );
      case 3:
        return (
          <Tag color='orange' size='large'>
            {t('管理')}
          </Tag>
        );
      case 4:
        return (
          <Tag color='purple' size='large'>
            {t('系统')}
          </Tag>
        );
      default:
        return (
          <Tag color='black' size='large'>
            {t('未知')}
          </Tag>
        );
    }
  }

  function renderIsStream(bool) {
    if (bool) {
      return (
        <Tag color='blue' size='large'>
          {t('流')}
        </Tag>
      );
    } else {
      return (
        <Tag color='purple' size='large'>
          {t('非流')}
        </Tag>
      );
    }
  }

  function renderUseTime(type) {
    const time = parseInt(type);
    if (time < 101) {
      return (
        <Tag color='green' size='large'>
          {' '}
          {time} s{' '}
        </Tag>
      );
    } else if (time < 300) {
      return (
        <Tag color='orange' size='large'>
          {' '}
          {time} s{' '}
        </Tag>
      );
    } else {
      return (
        <Tag color='red' size='large'>
          {' '}
          {time} s{' '}
        </Tag>
      );
    }
  }

  function renderFirstUseTime(type) {
    let time = parseFloat(type) / 1000.0;
    time = time.toFixed(1);
    if (time < 3) {
      return (
        <Tag color='green' size='large'>
          {' '}
          {time} s{' '}
        </Tag>
      );
    } else if (time < 10) {
      return (
        <Tag color='orange' size='large'>
          {' '}
          {time} s{' '}
        </Tag>
      );
    } else {
      return (
        <Tag color='red' size='large'>
          {' '}
          {time} s{' '}
        </Tag>
      );
    }
  }

  function renderModelName(record) {
    let other = getLogOther(record.other);
    let modelMapped =
      other?.is_model_mapped &&
      other?.upstream_model_name &&
      other?.upstream_model_name !== '';
    if (!modelMapped) {
      return (
        <Tag
          color={stringToColor(record.model_name)}
          size='large'
          onClick={(event) => {
            copyText(event, record.model_name).then((r) => {});
          }}
        >
          {' '}
          {record.model_name}{' '}
        </Tag>
      );
    } else {
      return (
        <>
          <Space vertical align={'start'}>
            <Popover
              content={
                <div style={{ padding: 10 }}>
                  <Space vertical align={'start'}>
                    <Tag
                      color={stringToColor(record.model_name)}
                      size='large'
                      onClick={(event) => {
                        copyText(event, record.model_name).then((r) => {});
                      }}
                    >
                      {t('请求并计费模型')} {record.model_name}{' '}
                    </Tag>
                    <Tag
                      color={stringToColor(other.upstream_model_name)}
                      size='large'
                      onClick={(event) => {
                        copyText(event, other.upstream_model_name).then(
                          (r) => {},
                        );
                      }}
                    >
                      {t('实际模型')} {other.upstream_model_name}{' '}
                    </Tag>
                  </Space>
                </div>
              }
            >
              <Tag
                color={stringToColor(record.model_name)}
                size='large'
                onClick={(event) => {
                  copyText(event, record.model_name).then((r) => {});
                }}
                suffixIcon={
                  <IconRefresh
                    style={{ width: '0.8em', height: '0.8em', opacity: 0.6 }}
                  />
                }
              >
                {' '}
                {record.model_name}{' '}
              </Tag>
            </Popover>
            {/*<Tooltip content={t('实际模型')}>*/}
            {/*  <Tag*/}
            {/*    color={stringToColor(other.upstream_model_name)}*/}
            {/*    size='large'*/}
            {/*    onClick={(event) => {*/}
            {/*      copyText(event, other.upstream_model_name).then(r => {});*/}
            {/*    }}*/}
            {/*  >*/}
            {/*    {' '}{other.upstream_model_name}{' '}*/}
            {/*  </Tag>*/}
            {/*</Tooltip>*/}
          </Space>
        </>
      );
    }
  }

  // Define column keys for selection
  const COLUMN_KEYS = {
    TIME: 'time',
    CHANNEL: 'channel',
    USERNAME: 'username',
    TOKEN: 'token',
    GROUP: 'group',
    TYPE: 'type',
    MODEL: 'model',
    USE_TIME: 'use_time',
    PROMPT: 'prompt',
    COMPLETION: 'completion',
    COST: 'cost',
    RETRY: 'retry',
    DETAILS: 'details',
  };

  // State for column visibility
  const [visibleColumns, setVisibleColumns] = useState({});
  const [showColumnSelector, setShowColumnSelector] = useState(false);

  // Load saved column preferences from localStorage
  useEffect(() => {
    const savedColumns = localStorage.getItem('logs-table-columns');
    if (savedColumns) {
      try {
        const parsed = JSON.parse(savedColumns);
        // Make sure all columns are accounted for
        const defaults = getDefaultColumnVisibility();
        const merged = { ...defaults, ...parsed };
        setVisibleColumns(merged);
      } catch (e) {
        console.error('Failed to parse saved column preferences', e);
        initDefaultColumns();
      }
    } else {
      initDefaultColumns();
    }
  }, []);

  // Get default column visibility based on user role
  const getDefaultColumnVisibility = () => {
    return {
      [COLUMN_KEYS.TIME]: true,
      [COLUMN_KEYS.CHANNEL]: isAdminUser,
      [COLUMN_KEYS.USERNAME]: isAdminUser,
      [COLUMN_KEYS.TOKEN]: true,
      [COLUMN_KEYS.GROUP]: true,
      [COLUMN_KEYS.TYPE]: true,
      [COLUMN_KEYS.MODEL]: true,
      [COLUMN_KEYS.USE_TIME]: true,
      [COLUMN_KEYS.PROMPT]: true,
      [COLUMN_KEYS.COMPLETION]: true,
      [COLUMN_KEYS.COST]: true,
      [COLUMN_KEYS.RETRY]: isAdminUser,
      [COLUMN_KEYS.DETAILS]: true,
    };
  };

  // Initialize default column visibility
  const initDefaultColumns = () => {
    const defaults = getDefaultColumnVisibility();
    setVisibleColumns(defaults);
    localStorage.setItem('logs-table-columns', JSON.stringify(defaults));
  };

  // Handle column visibility change
  const handleColumnVisibilityChange = (columnKey, checked) => {
    const updatedColumns = { ...visibleColumns, [columnKey]: checked };
    setVisibleColumns(updatedColumns);
  };

  // Handle "Select All" checkbox
  const handleSelectAll = (checked) => {
    const allKeys = Object.keys(COLUMN_KEYS).map((key) => COLUMN_KEYS[key]);
    const updatedColumns = {};

    allKeys.forEach((key) => {
      // For admin-only columns, only enable them if user is admin
      if (
        (key === COLUMN_KEYS.CHANNEL ||
          key === COLUMN_KEYS.USERNAME ||
          key === COLUMN_KEYS.RETRY) &&
        !isAdminUser
      ) {
        updatedColumns[key] = false;
      } else {
        updatedColumns[key] = checked;
      }
    });

    setVisibleColumns(updatedColumns);
  };

  // Define all columns
  const allColumns = [
    {
      key: COLUMN_KEYS.TIME,
      title: t('时间'),
      dataIndex: 'timestamp2string',
    },
    {
      key: COLUMN_KEYS.CHANNEL,
      title: t('渠道'),
      dataIndex: 'channel',
      className: isAdmin() ? 'tableShow' : 'tableHiddle',
      render: (text, record, index) => {
        return isAdminUser ? (
          record.type === 0 || record.type === 2 ? (
            <div>
              {
                <Tooltip content={record.channel_name || '[未知]'}>
                  <Tag
                    color={colors[parseInt(text) % colors.length]}
                    size='large'
                  >
                    {' '}
                    {text}{' '}
                  </Tag>
                </Tooltip>
              }
            </div>
          ) : (
            <></>
          )
        ) : (
          <></>
        );
      },
    },
    {
      key: COLUMN_KEYS.USERNAME,
      title: t('用户'),
      dataIndex: 'username',
      className: isAdmin() ? 'tableShow' : 'tableHiddle',
      render: (text, record, index) => {
        return isAdminUser ? (
          <div>
            <Avatar
              size='small'
              color={stringToColor(text)}
              style={{ marginRight: 4 }}
              onClick={(event) => {
                event.stopPropagation();
                showUserInfo(record.user_id);
              }}
            >
              {typeof text === 'string' && text.slice(0, 1)}
            </Avatar>
            {text}
          </div>
        ) : (
          <></>
        );
      },
    },
    {
      key: COLUMN_KEYS.TOKEN,
      title: t('令牌'),
      dataIndex: 'token_name',
      render: (text, record, index) => {
        return record.type === 0 || record.type === 2 ? (
          <div>
            <Tag
              color='grey'
              size='large'
              onClick={(event) => {
                //cancel the row click event
                copyText(event, text);
              }}
            >
              {' '}
              {t(text)}{' '}
            </Tag>
          </div>
        ) : (
          <></>
        );
      },
    },
    {
      key: COLUMN_KEYS.GROUP,
      title: t('分组'),
      dataIndex: 'group',
      render: (text, record, index) => {
        if (record.type === 0 || record.type === 2) {
          if (record.group) {
            return <>{renderGroup(record.group)}</>;
          } else {
            let other = null;
            try {
              other = JSON.parse(record.other);
            } catch (e) {
              console.error(
                `Failed to parse record.other: "${record.other}".`,
                e,
              );
            }
            if (other === null) {
              return <></>;
            }
            if (other.group !== undefined) {
              return <>{renderGroup(other.group)}</>;
            } else {
              return <></>;
            }
          }
        } else {
          return <></>;
        }
      },
    },
    {
      key: COLUMN_KEYS.TYPE,
      title: t('类型'),
      dataIndex: 'type',
      render: (text, record, index) => {
        return <>{renderType(text)}</>;
      },
    },
    {
      key: COLUMN_KEYS.MODEL,
      title: t('模型'),
      dataIndex: 'model_name',
      render: (text, record, index) => {
        return record.type === 0 || record.type === 2 ? (
          <>{renderModelName(record)}</>
        ) : (
          <></>
        );
      },
    },
    {
      key: COLUMN_KEYS.USE_TIME,
      title: t('用时/首字'),
      dataIndex: 'use_time',
      render: (text, record, index) => {
        if (record.is_stream) {
          let other = getLogOther(record.other);
          return (
            <>
              <Space>
                {renderUseTime(text)}
                {renderFirstUseTime(other?.frt)}
                {renderIsStream(record.is_stream)}
              </Space>
            </>
          );
        } else {
          return (
            <>
              <Space>
                {renderUseTime(text)}
                {renderIsStream(record.is_stream)}
              </Space>
            </>
          );
        }
      },
    },
    {
      key: COLUMN_KEYS.PROMPT,
      title: t('提示'),
      dataIndex: 'prompt_tokens',
      render: (text, record, index) => {
        return record.type === 0 || record.type === 2 ? (
          <>{<span> {text} </span>}</>
        ) : (
          <></>
        );
      },
    },
    {
      key: COLUMN_KEYS.COMPLETION,
      title: t('补全'),
      dataIndex: 'completion_tokens',
      render: (text, record, index) => {
        return parseInt(text) > 0 &&
          (record.type === 0 || record.type === 2) ? (
          <>{<span> {text} </span>}</>
        ) : (
          <></>
        );
      },
    },
    {
      key: COLUMN_KEYS.COST,
      title: t('花费'),
      dataIndex: 'quota',
      render: (text, record, index) => {
        return record.type === 0 || record.type === 2 ? (
          <>{renderQuota(text, 6)}</>
        ) : (
          <></>
        );
      },
    },
    {
      key: COLUMN_KEYS.RETRY,
      title: t('重试'),
      dataIndex: 'retry',
      className: isAdmin() ? 'tableShow' : 'tableHiddle',
      render: (text, record, index) => {
        let content = t('渠道') + `：${record.channel}`;
        if (record.other !== '') {
          let other = JSON.parse(record.other);
          if (other === null) {
            return <></>;
          }
          if (other.admin_info !== undefined) {
            if (
              other.admin_info.use_channel !== null &&
              other.admin_info.use_channel !== undefined &&
              other.admin_info.use_channel !== ''
            ) {
              // channel id array
              let useChannel = other.admin_info.use_channel;
              let useChannelStr = useChannel.join('->');
              content = t('渠道') + `：${useChannelStr}`;
            }
          }
        }
        return isAdminUser ? <div>{content}</div> : <></>;
      },
    },
    {
      key: COLUMN_KEYS.DETAILS,
      title: t('详情'),
      dataIndex: 'content',
      render: (text, record, index) => {
        let other = getLogOther(record.other);
        if (other == null || record.type !== 2) {
          return (
            <Paragraph
              ellipsis={{
                rows: 2,
                showTooltip: {
                  type: 'popover',
                  opts: { style: { width: 240 } },
                },
              }}
              style={{ maxWidth: 240 }}
            >
              {text}
            </Paragraph>
          );
        }

        let content = other?.claude
          ? renderClaudeModelPriceSimple(
              other.model_ratio,
              other.model_price,
              other.group_ratio,
              other.cache_tokens || 0,
              other.cache_ratio || 1.0,
              other.cache_creation_tokens || 0,
              other.cache_creation_ratio || 1.0,
            )
          : renderModelPriceSimple(
              other.model_ratio,
              other.model_price,
              other.group_ratio,
              other.cache_tokens || 0,
              other.cache_ratio || 1.0,
            );
        return (
          <Paragraph
            ellipsis={{
              rows: 2,
            }}
            style={{ maxWidth: 240 }}
          >
            {content}
          </Paragraph>
        );
      },
    },
  ];

  // Update table when column visibility changes
  useEffect(() => {
    if (Object.keys(visibleColumns).length > 0) {
      // Save to localStorage
      localStorage.setItem(
        'logs-table-columns',
        JSON.stringify(visibleColumns),
      );
    }
  }, [visibleColumns]);

  // Filter columns based on visibility settings
  const getVisibleColumns = () => {
    return allColumns.filter((column) => visibleColumns[column.key]);
  };

  // Column selector modal
  const renderColumnSelector = () => {
    return (
      <Modal
        title={t('列设置')}
        visible={showColumnSelector}
        onCancel={() => setShowColumnSelector(false)}
        footer={
          <>
            <Button onClick={() => initDefaultColumns()}>{t('重置')}</Button>
            <Button onClick={() => setShowColumnSelector(false)}>
              {t('取消')}
            </Button>
            <Button type='primary' onClick={() => setShowColumnSelector(false)}>
              {t('确定')}
            </Button>
          </>
        }
      >
        <div style={{ marginBottom: 20 }}>
          <Checkbox
            checked={Object.values(visibleColumns).every((v) => v === true)}
            indeterminate={
              Object.values(visibleColumns).some((v) => v === true) &&
              !Object.values(visibleColumns).every((v) => v === true)
            }
            onChange={(e) => handleSelectAll(e.target.checked)}
          >
            {t('全选')}
          </Checkbox>
        </div>
        <div
          style={{
            display: 'flex',
            flexWrap: 'wrap',
            maxHeight: '400px',
            overflowY: 'auto',
            border: '1px solid var(--semi-color-border)',
            borderRadius: '6px',
            padding: '16px',
          }}
        >
          {allColumns.map((column) => {
            // Skip admin-only columns for non-admin users
            if (
              !isAdminUser &&
              (column.key === COLUMN_KEYS.CHANNEL ||
                column.key === COLUMN_KEYS.USERNAME ||
                column.key === COLUMN_KEYS.RETRY)
            ) {
              return null;
            }

            return (
              <div
                key={column.key}
                style={{ width: '50%', marginBottom: 16, paddingRight: 8 }}
              >
                <Checkbox
                  checked={!!visibleColumns[column.key]}
                  onChange={(e) =>
                    handleColumnVisibilityChange(column.key, e.target.checked)
                  }
                >
                  {column.title}
                </Checkbox>
              </div>
            );
          })}
        </div>
      </Modal>
    );
  };

  const [styleState, styleDispatch] = useContext(StyleContext);
  const [logs, setLogs] = useState([]);
  const [expandData, setExpandData] = useState({});
  const [showStat, setShowStat] = useState(false);
  const [loading, setLoading] = useState(false);
  const [loadingStat, setLoadingStat] = useState(false);
  const [activePage, setActivePage] = useState(1);
  const [logCount, setLogCount] = useState(ITEMS_PER_PAGE);
  const [pageSize, setPageSize] = useState(ITEMS_PER_PAGE);
  const [logType, setLogType] = useState(0);
  const isAdminUser = isAdmin();
  let now = new Date();
  // 初始化start_timestamp为今天0点
  const [inputs, setInputs] = useState({
    username: '',
    token_name: '',
    model_name: '',
    start_timestamp: timestamp2string(getTodayStartTimestamp()),
    end_timestamp: timestamp2string(now.getTime() / 1000 + 3600),
    channel: '',
    group: '',
  });
  const {
    username,
    token_name,
    model_name,
    start_timestamp,
    end_timestamp,
    channel,
    group,
  } = inputs;

  const [stat, setStat] = useState({
    quota: 0,
    token: 0,
  });

  const handleInputChange = (value, name) => {
    setInputs((inputs) => ({ ...inputs, [name]: value }));
  };

  const getLogSelfStat = async () => {
    let localStartTimestamp = Date.parse(start_timestamp) / 1000;
    let localEndTimestamp = Date.parse(end_timestamp) / 1000;
    let url = `/api/log/self/stat?type=${logType}&token_name=${token_name}&model_name=${model_name}&start_timestamp=${localStartTimestamp}&end_timestamp=${localEndTimestamp}&group=${group}`;
    url = encodeURI(url);
    let res = await API.get(url);
    const { success, message, data } = res.data;
    if (success) {
      setStat(data);
    } else {
      showError(message);
    }
  };

  const getLogStat = async () => {
    let localStartTimestamp = Date.parse(start_timestamp) / 1000;
    let localEndTimestamp = Date.parse(end_timestamp) / 1000;
    let url = `/api/log/stat?type=${logType}&username=${username}&token_name=${token_name}&model_name=${model_name}&start_timestamp=${localStartTimestamp}&end_timestamp=${localEndTimestamp}&channel=${channel}&group=${group}`;
    url = encodeURI(url);
    let res = await API.get(url);
    const { success, message, data } = res.data;
    if (success) {
      setStat(data);
    } else {
      showError(message);
    }
  };

  const handleEyeClick = async () => {
    if (loadingStat) {
      return;
    }
    setLoadingStat(true);
    if (isAdminUser) {
      await getLogStat();
    } else {
      await getLogSelfStat();
    }
    setShowStat(true);
    setLoadingStat(false);
  };

  const showUserInfo = async (userId) => {
    if (!isAdminUser) {
      return;
    }
    const res = await API.get(`/api/user/${userId}`);
    const { success, message, data } = res.data;
    if (success) {
      Modal.info({
        title: t('用户信息'),
        content: (
          <div style={{ padding: 12 }}>
            <p>
              {t('用户名')}: {data.username}
            </p>
            <p>
              {t('余额')}: {renderQuota(data.quota)}
            </p>
            <p>
              {t('已用额度')}：{renderQuota(data.used_quota)}
            </p>
            <p>
              {t('请求次数')}：{renderNumber(data.request_count)}
            </p>
          </div>
        ),
        centered: true,
      });
    } else {
      showError(message);
    }
  };

  const setLogsFormat = (logs) => {
    let expandDatesLocal = {};
    for (let i = 0; i < logs.length; i++) {
      logs[i].timestamp2string = timestamp2string(logs[i].created_at);
      logs[i].key = logs[i].id;
      let other = getLogOther(logs[i].other);
      let expandDataLocal = [];
      if (isAdmin()) {
        // let content = '渠道：' + logs[i].channel;
        // if (other.admin_info !== undefined) {
        //   if (
        //     other.admin_info.use_channel !== null &&
        //     other.admin_info.use_channel !== undefined &&
        //     other.admin_info.use_channel !== ''
        //   ) {
        //     // channel id array
        //     let useChannel = other.admin_info.use_channel;
        //     let useChannelStr = useChannel.join('->');
        //     content = `渠道：${useChannelStr}`;
        //   }
        // }
        // expandDataLocal.push({
        //   key: '渠道重试',
        //   value: content,
        // })
      }
      if (isAdminUser && (logs[i].type === 0 || logs[i].type === 2)) {
        expandDataLocal.push({
          key: t('渠道信息'),
          value: `${logs[i].channel} - ${logs[i].channel_name || '[未知]'}`,
        });
      }
      if (other?.ws || other?.audio) {
        expandDataLocal.push({
          key: t('语音输入'),
          value: other.audio_input,
        });
        expandDataLocal.push({
          key: t('语音输出'),
          value: other.audio_output,
        });
        expandDataLocal.push({
          key: t('文字输入'),
          value: other.text_input,
        });
        expandDataLocal.push({
          key: t('文字输出'),
          value: other.text_output,
        });
      }
      if (other?.cache_tokens > 0) {
        expandDataLocal.push({
          key: t('缓存 Tokens'),
          value: other.cache_tokens,
        });
      }
      if (other?.cache_creation_tokens > 0) {
        expandDataLocal.push({
          key: t('缓存创建 Tokens'),
          value: other.cache_creation_tokens,
        });
      }
      if (logs[i].type === 2) {
        expandDataLocal.push({
          key: t('日志详情'),
          value: other?.claude
            ? renderClaudeLogContent(
                other?.model_ratio,
                other.completion_ratio,
                other.model_price,
                other.group_ratio,
                other.user_group_ratio,
                other.cache_ratio || 1.0,
                other.cache_creation_ratio || 1.0,
              )
            : renderLogContent(
                other?.model_ratio,
                other.completion_ratio,
                other.model_price,
                other.group_ratio,
                other.user_group_ratio,
              ),
        });
      }
      if (logs[i].type === 2) {
        let modelMapped =
          other?.is_model_mapped &&
          other?.upstream_model_name &&
          other?.upstream_model_name !== '';
        if (modelMapped) {
          expandDataLocal.push({
            key: t('请求并计费模型'),
            value: logs[i].model_name,
          });
          expandDataLocal.push({
            key: t('实际模型'),
            value: other.upstream_model_name,
          });
        }
        let content = '';
        if (other?.ws || other?.audio) {
          content = renderAudioModelPrice(
            other?.text_input,
            other?.text_output,
            other?.model_ratio,
            other?.model_price,
            other?.completion_ratio,
            other?.audio_input,
            other?.audio_output,
            other?.audio_ratio,
            other?.audio_completion_ratio,
            other?.group_ratio,
            other?.cache_tokens || 0,
            other?.cache_ratio || 1.0,
          );
        } else if (other?.claude) {
          content = renderClaudeModelPrice(
            logs[i].prompt_tokens,
            logs[i].completion_tokens,
            other.model_ratio,
            other.model_price,
            other.completion_ratio,
            other.group_ratio,
            other.cache_tokens || 0,
            other.cache_ratio || 1.0,
            other.cache_creation_tokens || 0,
            other.cache_creation_ratio || 1.0,
          );
        } else {
          content = renderModelPrice(
            logs[i].prompt_tokens,
            logs[i].completion_tokens,
            other?.model_ratio,
            other?.model_price,
            other?.completion_ratio,
            other?.group_ratio,
            other?.cache_tokens || 0,
            other?.cache_ratio || 1.0,
          );
        }
        expandDataLocal.push({
          key: t('计费过程'),
          value: content,
        });
        if (other?.reasoning_effort) {
          expandDataLocal.push({
            key: t('Reasoning Effort'),
            value: other.reasoning_effort,
          });
        }
      }
      expandDatesLocal[logs[i].key] = expandDataLocal;
    }

    setExpandData(expandDatesLocal);
    setLogs(logs);
  };

  const loadLogs = async (startIdx, pageSize, logType = 0) => {
    setLoading(true);

    let url = '';
    let localStartTimestamp = Date.parse(start_timestamp) / 1000;
    let localEndTimestamp = Date.parse(end_timestamp) / 1000;
    if (isAdminUser) {
      url = `/api/log/?p=${startIdx}&page_size=${pageSize}&type=${logType}&username=${username}&token_name=${token_name}&model_name=${model_name}&start_timestamp=${localStartTimestamp}&end_timestamp=${localEndTimestamp}&channel=${channel}&group=${group}`;
    } else {
      url = `/api/log/self/?p=${startIdx}&page_size=${pageSize}&type=${logType}&token_name=${token_name}&model_name=${model_name}&start_timestamp=${localStartTimestamp}&end_timestamp=${localEndTimestamp}&group=${group}`;
    }
    url = encodeURI(url);
    const res = await API.get(url);
    const { success, message, data } = res.data;
    if (success) {
      const newPageData = data.items;
      setActivePage(data.page);
      setPageSize(data.page_size);
      setLogCount(data.total);

      setLogsFormat(newPageData);
    } else {
      showError(message);
    }
    setLoading(false);
  };

  const handlePageChange = (page) => {
    setActivePage(page);
    loadLogs(page, pageSize, logType).then((r) => {});
  };

  const handlePageSizeChange = async (size) => {
    localStorage.setItem('page-size', size + '');
    setPageSize(size);
    setActivePage(1);
    loadLogs(activePage, size)
      .then()
      .catch((reason) => {
        showError(reason);
      });
  };

  const refresh = async () => {
    setActivePage(1);
    handleEyeClick();
    await loadLogs(activePage, pageSize, logType);
  };

  const copyText = async (e, text) => {
    e.stopPropagation();
    if (await copy(text)) {
      showSuccess('已复制：' + text);
    } else {
      Modal.error({ title: t('无法复制到剪贴板，请手动复制'), content: text });
    }
  };

  useEffect(() => {
    const localPageSize =
      parseInt(localStorage.getItem('page-size')) || ITEMS_PER_PAGE;
    setPageSize(localPageSize);
    loadLogs(activePage, localPageSize)
      .then()
      .catch((reason) => {
        showError(reason);
      });
    handleEyeClick();
  }, []);

  const expandRowRender = (record, index) => {
    return <Descriptions data={expandData[record.key]} />;
  };

  return (
    <>
      {renderColumnSelector()}
      <Layout>
        <Header>
          <Spin spinning={loadingStat}>
            <Space>
              <Tag
                color='blue'
                size='large'
                style={{
                  padding: 15,
                  borderRadius: '8px',
                  fontWeight: 500,
                  boxShadow: '0 2px 8px rgba(0, 0, 0, 0.1)',
                }}
              >
                {t('消耗额度')}: {renderQuota(stat.quota)}
              </Tag>
              <Tag
                color='pink'
                size='large'
                style={{
                  padding: 15,
                  borderRadius: '8px',
                  fontWeight: 500,
                  boxShadow: '0 2px 8px rgba(0, 0, 0, 0.1)',
                }}
              >
                RPM: {stat.rpm}
              </Tag>
              <Tag
                color='white'
                size='large'
                style={{
                  padding: 15,
                  border: 'none',
                  boxShadow: '0 2px 8px rgba(0, 0, 0, 0.1)',
                  borderRadius: '8px',
                  fontWeight: 500,
                }}
              >
                TPM: {stat.tpm}
              </Tag>
            </Space>
          </Spin>
        </Header>
        <Form layout='horizontal' style={{ marginTop: 10 }}>
          <>
            <Form.Section>
              <div style={{ marginBottom: 10 }}>
                {styleState.isMobile ? (
                  <div>
                    <Form.DatePicker
                      field='start_timestamp'
                      label={t('起始时间')}
                      style={{ width: 272 }}
                      initValue={start_timestamp}
                      type='dateTime'
                      onChange={(value) => {
                        console.log(value);
                        handleInputChange(value, 'start_timestamp');
                      }}
                    />
                    <Form.DatePicker
                      field='end_timestamp'
                      fluid
                      label={t('结束时间')}
                      style={{ width: 272 }}
                      initValue={end_timestamp}
                      type='dateTime'
                      onChange={(value) =>
                        handleInputChange(value, 'end_timestamp')
                      }
                    />
                  </div>
                ) : (
                  <Form.DatePicker
                    field='range_timestamp'
                    label={t('时间范围')}
                    initValue={[start_timestamp, end_timestamp]}
                    type='dateTimeRange'
                    name='range_timestamp'
                    onChange={(value) => {
                      if (Array.isArray(value) && value.length === 2) {
                        handleInputChange(value[0], 'start_timestamp');
                        handleInputChange(value[1], 'end_timestamp');
                      }
                    }}
                  />
                )}
              </div>
            </Form.Section>
            <Form.Input
              field='token_name'
              label={t('令牌名称')}
              value={token_name}
              placeholder={t('可选值')}
              name='token_name'
              onChange={(value) => handleInputChange(value, 'token_name')}
            />
            <Form.Input
              field='model_name'
              label={t('模型名称')}
              value={model_name}
              placeholder={t('可选值')}
              name='model_name'
              onChange={(value) => handleInputChange(value, 'model_name')}
            />
            <Form.Input
              field='group'
              label={t('分组')}
              value={group}
              placeholder={t('可选值')}
              name='group'
              onChange={(value) => handleInputChange(value, 'group')}
            />
            {isAdminUser && (
              <>
                <Form.Input
                  field='channel'
                  label={t('渠道 ID')}
                  value={channel}
                  placeholder={t('可选值')}
                  name='channel'
                  onChange={(value) => handleInputChange(value, 'channel')}
                />
                <Form.Input
                  field='username'
                  label={t('用户名称')}
                  value={username}
                  placeholder={t('可选值')}
                  name='username'
                  onChange={(value) => handleInputChange(value, 'username')}
                />
              </>
            )}
            <Button
              label={t('查询')}
              type='primary'
              htmlType='submit'
              className='btn-margin-right'
              onClick={refresh}
              loading={loading}
              style={{ marginTop: 24 }}
            >
              {t('查询')}
            </Button>
            <Form.Section></Form.Section>
          </>
        </Form>
        <div style={{ marginTop: 10 }}>
          <Select
            defaultValue='0'
            style={{ width: 120 }}
            onChange={(value) => {
              setLogType(parseInt(value));
              loadLogs(0, pageSize, parseInt(value));
            }}
          >
            <Select.Option value='0'>{t('全部')}</Select.Option>
            <Select.Option value='1'>{t('充值')}</Select.Option>
            <Select.Option value='2'>{t('消费')}</Select.Option>
            <Select.Option value='3'>{t('管理')}</Select.Option>
            <Select.Option value='4'>{t('系统')}</Select.Option>
          </Select>
          <Button
            theme='light'
            type='tertiary'
            icon={<IconSetting />}
            onClick={() => setShowColumnSelector(true)}
            style={{ marginLeft: 8 }}
          >
            {t('列设置')}
          </Button>
        </div>
        <Table
          style={{ marginTop: 5 }}
          columns={getVisibleColumns()}
          expandedRowRender={expandRowRender}
          expandRowByClick={true}
          dataSource={logs}
          rowKey='key'
          pagination={{
            formatPageText: (page) =>
              t('第 {{start}} - {{end}} 条，共 {{total}} 条', {
                start: page.currentStart,
                end: page.currentEnd,
                total: logCount,
              }),
            currentPage: activePage,
            pageSize: pageSize,
            total: logCount,
            pageSizeOpts: [10, 20, 50, 100],
            showSizeChanger: true,
            onPageSizeChange: (size) => {
              handlePageSizeChange(size);
            },
            onPageChange: handlePageChange,
          }}
        />
      </Layout>
    </>
  );
};

export default LogsTable;
