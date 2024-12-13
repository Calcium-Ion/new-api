import React, { useEffect, useState } from 'react';
import {
  API,
  copy,
  showError,
  showSuccess,
  timestamp2string,
} from '../helpers';

import { ITEMS_PER_PAGE } from '../constants';
import {renderGroup, renderQuota} from '../helpers/render';
import {
  Button, Divider,
  Dropdown,
  Form,
  Modal,
  Popconfirm,
  Popover, Space,
  SplitButtonGroup,
  Table,
  Tag,
} from '@douyinfe/semi-ui';

import { IconTreeTriangleDown } from '@douyinfe/semi-icons';
import EditToken from '../pages/Token/EditToken';
import { useTranslation } from 'react-i18next';

function renderTimestamp(timestamp) {
  return <>{timestamp2string(timestamp)}</>;
}

const TokensTable = () => {

  const { t } = useTranslation();

  const renderStatus = (status, model_limits_enabled = false) => {
    switch (status) {
      case 1:
        if (model_limits_enabled) {
          return (
            <Tag color='green' size='large'>
              {t('已启用：限制模型')}
            </Tag>
          );
        } else {
          return (
            <Tag color='green' size='large'>
              {t('已启用')}
            </Tag>
          );
        }
      case 2:
        return (
          <Tag color='red' size='large'>
            {t('已禁用')}
          </Tag>
        );
      case 3:
        return (
          <Tag color='yellow' size='large'>
            {t('已过期')}
          </Tag>
        );
      case 4:
        return (
          <Tag color='grey' size='large'>
            {t('已耗尽')}
          </Tag>
        );
      default:
        return (
          <Tag color='black' size='large'>
            {t('未知状态')}
          </Tag>
        );
    }
  };

  const columns = [
    {
      title: t('名称'),
      dataIndex: 'name',
    },
    {
      title: t('状态'),
      dataIndex: 'status',
      key: 'status',
      render: (text, record, index) => {
        return <div>
          <Space>
            {renderStatus(text, record.model_limits_enabled)}
            {renderGroup(record.group)}
          </Space>
        </div>;
      },
    },
    {
      title: t('已用额度'),
      dataIndex: 'used_quota',
      render: (text, record, index) => {
        return <div>{renderQuota(parseInt(text))}</div>;
      },
    },
    {
      title: t('剩余额度'),
      dataIndex: 'remain_quota',
      render: (text, record, index) => {
        return (
          <div>
            {record.unlimited_quota ? (
              <Tag size={'large'} color={'white'}>
                {t('无限制')}
              </Tag>
            ) : (
              <Tag size={'large'} color={'light-blue'}>
                {renderQuota(parseInt(text))}
              </Tag>
            )}
          </div>
        );
      },
    },
    {
      title: t('创建时间'),
      dataIndex: 'created_time',
      render: (text, record, index) => {
        return <div>{renderTimestamp(text)}</div>;
      },
    },
    {
      title: t('过期时间'),
      dataIndex: 'expired_time',
      render: (text, record, index) => {
        return (
          <div>
            {record.expired_time === -1 ? t('永不过期') : renderTimestamp(text)}
          </div>
        );
      },
    },
    {
      title: '',
      dataIndex: 'operate',
      render: (text, record, index) => {
        let chats = localStorage.getItem('chats');
        let chatsArray = []
        let chatLink = localStorage.getItem('chat_link');
        let mjLink = localStorage.getItem('chat_link2');
        let shouldUseCustom = true;
        if (chatLink) {
          shouldUseCustom = false;
          chatLink += `/#/?settings={"key":"{key}","url":"{address}"}`;
          chatsArray.push({
            node: 'item',
            key: 'default',
            name: 'ChatGPT Next Web',
            onClick: () => {
              onOpenLink('default', chatLink, record);
            },
          });
        }
        if (mjLink) {
          shouldUseCustom = false;
          mjLink += `/#/?settings={"key":"{key}","url":"{address}"}`;
          chatsArray.push({
            node: 'item',
            key: 'mj',
            name: 'ChatGPT Next Midjourney',
            onClick: () => {
              onOpenLink('mj', mjLink, record);
            },
          });
        }
        if (shouldUseCustom) {
          try {
            // console.log(chats);
            chats = JSON.parse(chats);
            // check chats is array
            if (Array.isArray(chats)) {
              for (let i = 0; i < chats.length; i++) {
                let chat = {}
                chat.node = 'item';
                // c is a map
                // chat.key = chats[i].name;
                // console.log(chats[i])
                for (let key in chats[i]) {
                  if (chats[i].hasOwnProperty(key)) {
                    chat.key = i;
                    chat.name = key;
                    chat.onClick = () => {
                      onOpenLink(key, chats[i][key], record);
                    }
                  }
                }
                chatsArray.push(chat);
              }
            }

          } catch (e) {
            console.log(e);
            showError(t('聊天链接配置错误，请联系管理员'));
          }
        }
        return (
          <div>
            <Popover
              content={'sk-' + record.key}
              style={{ padding: 20 }}
              position='top'
            >
              <Button theme='light' type='tertiary' style={{ marginRight: 1 }}>
                {t('查看')}
              </Button>
            </Popover>
            <Button
              theme='light'
              type='secondary'
              style={{ marginRight: 1 }}
              onClick={async (text) => {
                await copyText('sk-' + record.key);
              }}
            >
              {t('复制')}
            </Button>
            <SplitButtonGroup
              style={{ marginRight: 1 }}
              aria-label={t('项目操作按钮组')}
            >
              <Button
                theme='light'
                style={{ color: 'rgba(var(--semi-teal-7), 1)' }}
                onClick={() => {
                  if (chatsArray.length === 0) {
                    showError(t('请联系管理员配置聊天链接'));
                  } else {
                    onOpenLink('default', chats[0][Object.keys(chats[0])[0]], record);
                  }
                }}
              >
                {t('聊天')}
              </Button>
              <Dropdown
                trigger='click'
                position='bottomRight'
                menu={chatsArray}
              >
                <Button
                  style={{
                    padding: '8px 4px',
                    color: 'rgba(var(--semi-teal-7), 1)',
                  }}
                  type='primary'
                  icon={<IconTreeTriangleDown />}
                ></Button>
              </Dropdown>
            </SplitButtonGroup>
            <Popconfirm
              title={t('确定是否要删除此令牌？')}
              content={t('此修改将不可逆')}
              okType={'danger'}
              position={'left'}
              onConfirm={() => {
                manageToken(record.id, 'delete', record).then(() => {
                  removeRecord(record.key);
                });
              }}
            >
              <Button theme='light' type='danger' style={{ marginRight: 1 }}>
                {t('删除')}
              </Button>
            </Popconfirm>
            {record.status === 1 ? (
              <Button
                theme='light'
                type='warning'
                style={{ marginRight: 1 }}
                onClick={async () => {
                  manageToken(record.id, 'disable', record);
                }}
              >
                {t('禁用')}
              </Button>
            ) : (
              <Button
                theme='light'
                type='secondary'
                style={{ marginRight: 1 }}
                onClick={async () => {
                  manageToken(record.id, 'enable', record);
                }}
              >
                {t('启用')}
              </Button>
            )}
            <Button
              theme='light'
              type='tertiary'
              style={{ marginRight: 1 }}
              onClick={() => {
                setEditingToken(record);
                setShowEdit(true);
              }}
            >
              {t('编辑')}
            </Button>
          </div>
        );
      },
    },
  ];

  const [pageSize, setPageSize] = useState(ITEMS_PER_PAGE);
  const [showEdit, setShowEdit] = useState(false);
  const [tokens, setTokens] = useState([]);
  const [selectedKeys, setSelectedKeys] = useState([]);
  const [tokenCount, setTokenCount] = useState(pageSize);
  const [loading, setLoading] = useState(true);
  const [activePage, setActivePage] = useState(1);
  const [searchKeyword, setSearchKeyword] = useState('');
  const [searchToken, setSearchToken] = useState('');
  const [searching, setSearching] = useState(false);
  const [chats, setChats] = useState([]);
  const [editingToken, setEditingToken] = useState({
    id: undefined,
  });

  const closeEdit = () => {
    setShowEdit(false);
    setTimeout(() => {
      setEditingToken({
        id: undefined,
      });
    }, 500);
  };

  const setTokensFormat = (tokens) => {
    setTokens(tokens);
    if (tokens.length >= pageSize) {
      setTokenCount(tokens.length + pageSize);
    } else {
      setTokenCount(tokens.length);
    }
  };

  let pageData = tokens.slice(
    (activePage - 1) * pageSize,
    activePage * pageSize,
  );
  const loadTokens = async (startIdx) => {
    setLoading(true);
    const res = await API.get(`/api/token/?p=${startIdx}&size=${pageSize}`);
    const { success, message, data } = res.data;
    if (success) {
      if (startIdx === 0) {
        setTokensFormat(data);
      } else {
        let newTokens = [...tokens];
        newTokens.splice(startIdx * pageSize, data.length, ...data);
        setTokensFormat(newTokens);
      }
    } else {
      showError(message);
    }
    setLoading(false);
  };

  const refresh = async () => {
    await loadTokens(activePage - 1);
  };

  const copyText = async (text) => {
    if (await copy(text)) {
      showSuccess(t('已复制到剪贴板！'));
    } else {
      Modal.error({
        title: t('无法复制到剪贴板，请手动复制'),
        content: text,
        size: 'large',
      });
    }
  };

  const onOpenLink = async (type, url, record) => {
    // console.log(type, url, key);
    let status = localStorage.getItem('status');
    let serverAddress = '';
    if (status) {
      status = JSON.parse(status);
      serverAddress = status.server_address;
    }
    if (serverAddress === '') {
      serverAddress = window.location.origin;
    }
    let encodedServerAddress = encodeURIComponent(serverAddress);
    url = url.replaceAll('{address}', encodedServerAddress);
    url = url.replaceAll('{key}', 'sk-' + record.key);

    window.open(url, '_blank');
  };

  useEffect(() => {
    loadTokens(0)
      .then()
      .catch((reason) => {
        showError(reason);
      });
  }, [pageSize]);

  const removeRecord = (key) => {
    let newDataSource = [...tokens];
    if (key != null) {
      let idx = newDataSource.findIndex((data) => data.key === key);

      if (idx > -1) {
        newDataSource.splice(idx, 1);
        setTokensFormat(newDataSource);
      }
    }
  };

  const manageToken = async (id, action, record) => {
    setLoading(true);
    let data = { id };
    let res;
    switch (action) {
      case 'delete':
        res = await API.delete(`/api/token/${id}/`);
        break;
      case 'enable':
        data.status = 1;
        res = await API.put('/api/token/?status_only=true', data);
        break;
      case 'disable':
        data.status = 2;
        res = await API.put('/api/token/?status_only=true', data);
        break;
    }
    const { success, message } = res.data;
    if (success) {
      showSuccess('操作成功完成！');
      let token = res.data.data;
      let newTokens = [...tokens];
      // let realIdx = (activePage - 1) * ITEMS_PER_PAGE + idx;
      if (action === 'delete') {
      } else {
        record.status = token.status;
        // newTokens[realIdx].status = token.status;
      }
      setTokensFormat(newTokens);
    } else {
      showError(message);
    }
    setLoading(false);
  };

  const searchTokens = async () => {
    if (searchKeyword === '' && searchToken === '') {
      // if keyword is blank, load files instead.
      await loadTokens(0);
      setActivePage(1);
      return;
    }
    setSearching(true);
    const res = await API.get(
      `/api/token/search?keyword=${searchKeyword}&token=${searchToken}`,
    );
    const { success, message, data } = res.data;
    if (success) {
      setTokensFormat(data);
      setActivePage(1);
    } else {
      showError(message);
    }
    setSearching(false);
  };

  const handleKeywordChange = async (value) => {
    setSearchKeyword(value.trim());
  };

  const handleSearchTokenChange = async (value) => {
    setSearchToken(value.trim());
  };

  const sortToken = (key) => {
    if (tokens.length === 0) return;
    setLoading(true);
    let sortedTokens = [...tokens];
    sortedTokens.sort((a, b) => {
      return ('' + a[key]).localeCompare(b[key]);
    });
    if (sortedTokens[0].id === tokens[0].id) {
      sortedTokens.reverse();
    }
    setTokens(sortedTokens);
    setLoading(false);
  };

  const handlePageChange = (page) => {
    setActivePage(page);
    if (page === Math.ceil(tokens.length / pageSize) + 1) {
      // In this case we have to load more data and then append them.
      loadTokens(page - 1).then((r) => {});
    }
  };

  const rowSelection = {
    onSelect: (record, selected) => {},
    onSelectAll: (selected, selectedRows) => {},
    onChange: (selectedRowKeys, selectedRows) => {
      setSelectedKeys(selectedRows);
    },
  };

  const handleRow = (record, index) => {
    if (record.status !== 1) {
      return {
        style: {
          background: 'var(--semi-color-disabled-border)',
        },
      };
    } else {
      return {};
    }
  };

  return (
    <>
      <EditToken
        refresh={refresh}
        editingToken={editingToken}
        visiable={showEdit}
        handleClose={closeEdit}
      ></EditToken>
      <Form
        layout='horizontal'
        style={{ marginTop: 10 }}
        labelPosition={'left'}
      >
        <Form.Input
          field='keyword'
          label={t('搜索关键字')}
          placeholder={t('令牌名称')}
          value={searchKeyword}
          loading={searching}
          onChange={handleKeywordChange}
        />
        <Form.Input
          field='token'
          label={t('密钥')}
          placeholder={t('密钥')}
          value={searchToken}
          loading={searching}
          onChange={handleSearchTokenChange}
        />
        <Button
          label={t('查询')}
          type='primary'
          htmlType='submit'
          className='btn-margin-right'
          onClick={searchTokens}
          style={{ marginRight: 8 }}
        >
          {t('查询')}
        </Button>
      </Form>
      <Divider style={{margin:'15px 0'}}/>
      <div>
        <Button
            theme='light'
            type='primary'
            style={{ marginRight: 8 }}
            onClick={() => {
              setEditingToken({
                id: undefined,
              });
              setShowEdit(true);
            }}
        >
            {t('添加令牌')}
        </Button>
        <Button
            label={t('复制所选令牌')}
            type='warning'
            onClick={async () => {
              if (selectedKeys.length === 0) {
                showError(t('请至少选择一个令牌！'));
                return;
              }
              let keys = '';
              for (let i = 0; i < selectedKeys.length; i++) {
                keys +=
                    selectedKeys[i].name + '    sk-' + selectedKeys[i].key + '\n';
              }
              await copyText(keys);
            }}
        >
          {t('复制所选令牌到剪贴板')}
        </Button>
      </div>

      <Table
        style={{ marginTop: 20 }}
        columns={columns}
        dataSource={pageData}
        pagination={{
          currentPage: activePage,
          pageSize: pageSize,
          total: tokenCount,
          showSizeChanger: true,
          pageSizeOptions: [10, 20, 50, 100],
          formatPageText: (page) =>
            t('第 {{start}} - {{end}} 条，共 {{total}} 条', {
              start: page.currentStart,
              end: page.currentEnd,
              total: tokens.length
            }),
          onPageSizeChange: (size) => {
            setPageSize(size);
            setActivePage(1);
          },
          onPageChange: handlePageChange,
        }}
        loading={loading}
        rowSelection={rowSelection}
        onRow={handleRow}
      ></Table>
    </>
  );
};

export default TokensTable;
