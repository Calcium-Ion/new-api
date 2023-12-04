import React, {useEffect, useState} from 'react';
import {Link} from 'react-router-dom';
import {API, copy, isAdmin, showError, showSuccess, showWarning, timestamp2string} from '../helpers';

import {ITEMS_PER_PAGE} from '../constants';
import {renderQuota, stringToColor} from '../helpers/render';
import {Avatar, Tag, Table, Button, Popover, Form, Modal, Popconfirm} from "@douyinfe/semi-ui";
import EditToken from "../pages/Token/EditToken";

const {Column} = Table;

const COPY_OPTIONS = [
    {key: 'next', text: 'ChatGPT Next Web', value: 'next'},
    {key: 'ama', text: 'AMA 问天', value: 'ama'},
    {key: 'opencat', text: 'OpenCat', value: 'opencat'},
];

const OPEN_LINK_OPTIONS = [
    {key: 'ama', text: 'AMA 问天', value: 'ama'},
    {key: 'opencat', text: 'OpenCat', value: 'opencat'},
];

function renderTimestamp(timestamp) {
    return (
        <>
            {timestamp2string(timestamp)}
        </>
    );
}

function renderStatus(status) {
    switch (status) {
        case 1:
            return <Tag color='green' size='large'>已启用</Tag>;
        case 2:
            return <Tag color='red' size='large'> 已禁用 </Tag>;
        case 3:
            return <Tag color='yellow' size='large'> 已过期 </Tag>;
        case 4:
            return <Tag color='grey' size='large'> 已耗尽 </Tag>;
        default:
            return <Tag color='black' size='large'> 未知状态 </Tag>;
    }
}

const TokensTable = () => {
    const columns = [
        {
            title: '名称',
            dataIndex: 'name',
        },
        {
            title: '状态',
            dataIndex: 'status',
            key: 'status',
            render: (text, record, index) => {
                return (
                    <div>
                        {renderStatus(text)}
                    </div>
                );
            },
        },
        {
            title: '已用额度',
            dataIndex: 'used_quota',
            render: (text, record, index) => {
                return (
                    <div>
                        {renderQuota(parseInt(text))}
                    </div>
                );
            },
        },
        {
            title: '剩余额度',
            dataIndex: 'remain_quota',
            render: (text, record, index) => {
                return (
                    <div>
                        {record.unlimited_quota ? <Tag size={'large'} color={'white'}>无限制</Tag> : <Tag size={'large'} color={'light-blue'}>{renderQuota(parseInt(text))}</Tag>}
                    </div>
                );
            },
        },
        {
            title: '创建时间',
            dataIndex: 'created_time',
            render: (text, record, index) => {
                return (
                    <div>
                        {renderTimestamp(text)}
                    </div>
                );
            },
        },
        {
            title: '过期时间',
            dataIndex: 'expired_time',
            render: (text, record, index) => {
                return (
                    <div>
                        {record.expired_time === -1 ? "永不过期" : renderTimestamp(text)}
                    </div>
                );
            },
        },
        {
            title: '',
            dataIndex: 'operate',
            render: (text, record, index) => (
                <div>
                    <Popover
                        content={
                            'sk-' + record.key
                        }
                        style={{padding: 20}}
                        position="top"
                    >
                        <Button theme='light' type='tertiary' style={{marginRight: 1}}>查看</Button>
                    </Popover>
                    <Button theme='light' type='secondary' style={{marginRight: 1}}
                            onClick={async (text) => {
                                await copyText('sk-' + record.key)
                            }}
                    >复制</Button>
                    <Popconfirm
                        title="确定是否要删除此令牌？"
                        content="此修改将不可逆"
                        okType={'danger'}
                        position={'left'}
                        onConfirm={() => {
                            manageToken(record.id, 'delete', record).then(
                                () => {
                                    removeRecord(record.key);
                                }
                            )
                        }}
                    >
                        <Button theme='light' type='danger' style={{marginRight: 1}}>删除</Button>
                    </Popconfirm>
                    {
                        record.status === 1 ?
                            <Button theme='light' type='warning' style={{marginRight: 1}} onClick={
                                async () => {
                                    manageToken(
                                        record.id,
                                        'disable',
                                        record
                                    )
                                }
                            }>禁用</Button> :
                            <Button theme='light' type='secondary' style={{marginRight: 1}} onClick={
                                async () => {
                                    manageToken(
                                        record.id,
                                        'enable',
                                        record
                                    );
                                }
                            }>启用</Button>
                    }
                    <Button theme='light' type='tertiary' style={{marginRight: 1}} onClick={
                        () => {
                            setEditingToken(record);
                            setShowEdit(true);
                        }
                    }>编辑</Button>
                </div>
            ),
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
    const [showTopUpModal, setShowTopUpModal] = useState(false);
    const [targetTokenIdx, setTargetTokenIdx] = useState(0);
    const [editingToken, setEditingToken] = useState({
        id: undefined,
    });

    const closeEdit = () => {
        setShowEdit(false);
    }

    const setTokensFormat = (tokens) => {
        setTokens(tokens);
        if (tokens.length >= pageSize) {
            setTokenCount(tokens.length + pageSize);
        } else {
            setTokenCount(tokens.length);
        }
    }

    let pageData = tokens.slice((activePage - 1) * pageSize, activePage * pageSize);
    const loadTokens = async (startIdx) => {
        setLoading(true);
        const res = await API.get(`/api/token/?p=${startIdx}&size=${pageSize}`);
        const {success, message, data} = res.data;
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

    const onPaginationChange = (e, {activePage}) => {
        (async () => {
            if (activePage === Math.ceil(tokens.length / pageSize) + 1) {
                // In this case we have to load more data and then append them.
                await loadTokens(activePage - 1);
            }
            setActivePage(activePage);
        })();
    };

    const refresh = async () => {
        await loadTokens(activePage - 1);
    };

    const onCopy = async (type, key) => {
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
        const nextLink = localStorage.getItem('chat_link');
        let nextUrl;

        if (nextLink) {
            nextUrl = nextLink + `/#/?settings={"key":"sk-${key}","url":"${serverAddress}"}`;
        } else {
            nextUrl = `https://chat.oneapi.pro/#/?settings={"key":"sk-${key}","url":"${serverAddress}"}`;
        }

        let url;
        switch (type) {
            case 'ama':
                url = `ama://set-api-key?server=${encodedServerAddress}&key=sk-${key}`;
                break;
            case 'opencat':
                url = `opencat://team/join?domain=${encodedServerAddress}&token=sk-${key}`;
                break;
            case 'next':
                url = nextUrl;
                break;
            default:
                url = `sk-${key}`;
        }
        // if (await copy(url)) {
        //     showSuccess('已复制到剪贴板！');
        // } else {
        //     showWarning('无法复制到剪贴板，请手动复制，已将令牌填入搜索框。');
        //     setSearchKeyword(url);
        // }
    };

    const copyText = async (text) => {
        if (await copy(text)) {
            showSuccess('已复制到剪贴板！');
        } else {
            // setSearchKeyword(text);
            Modal.error({ title: '无法复制到剪贴板，请手动复制', content: text });
        }
    }

    const onOpenLink = async (type, key) => {
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
        const chatLink = localStorage.getItem('chat_link');
        let defaultUrl;

        if (chatLink) {
            defaultUrl = chatLink + `/#/?settings={"key":"sk-${key}","url":"${serverAddress}"}`;
        } else {
            defaultUrl = `https://chat.oneapi.pro/#/?settings={"key":"sk-${key}","url":"${serverAddress}"}`;
        }
        let url;
        switch (type) {
            case 'ama':
                url = `ama://set-api-key?server=${encodedServerAddress}&key=sk-${key}`;
                break;

            case 'opencat':
                url = `opencat://team/join?domain=${encodedServerAddress}&token=sk-${key}`;
                break;

            default:
                url = defaultUrl;
        }

        window.open(url, '_blank');
    }

    useEffect(() => {
        loadTokens(0)
            .then()
            .catch((reason) => {
                showError(reason);
            });
    }, [pageSize]);

    const removeRecord = key => {
        let newDataSource = [...tokens];
        if (key != null) {
            let idx = newDataSource.findIndex(data => data.key === key);

            if (idx > -1) {
                newDataSource.splice(idx, 1);
                setTokensFormat(newDataSource);
            }
        }
    };

    const manageToken = async (id, action, record) => {
        setLoading(true);
        let data = {id};
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
        const {success, message} = res.data;
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
        const res = await API.get(`/api/token/search?keyword=${searchKeyword}&token=${searchToken}`);
        const {success, message, data} = res.data;
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


    const handlePageChange = page => {
        setActivePage(page);
        if (page === Math.ceil(tokens.length / pageSize) + 1) {
            // In this case we have to load more data and then append them.
            loadTokens(page - 1).then(r => {
            });
        }
    };

    const rowSelection = {
        onSelect: (record, selected) => {
        },
        onSelectAll: (selected, selectedRows) => {
        },
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
            <EditToken refresh={refresh} editingToken={editingToken} visiable={showEdit} handleClose={closeEdit}></EditToken>
            <Form layout='horizontal' style={{marginTop: 10}} labelPosition={'left'}>
                <Form.Input
                    field="keyword"
                    label='搜索关键字'
                    placeholder='令牌名称'
                    value={searchKeyword}
                    loading={searching}
                    onChange={handleKeywordChange}
                />
                <Form.Input
                    field="token"
                    label='Key'
                    placeholder='密钥'
                    value={searchToken}
                    loading={searching}
                    onChange={handleSearchTokenChange}
                />
                <Button label='查询' type="primary" htmlType="submit" className="btn-margin-right"
                        onClick={searchTokens} style={{marginRight: 8}}>查询</Button>
            </Form>

            <Table style={{marginTop: 20}} columns={columns} dataSource={pageData} pagination={{
                currentPage: activePage,
                pageSize: pageSize,
                total: tokenCount,
                showSizeChanger: true,
                pageSizeOptions: [10, 20, 50, 100],
                formatPageText: (page) => `第 ${page.currentStart} - ${page.currentEnd} 条，共 ${tokens.length} 条`,
                onPageSizeChange: (size) => {
                    setPageSize(size);
                    setActivePage(1);
                },
                onPageChange: handlePageChange,
            }} loading={loading} rowSelection={rowSelection} onRow={handleRow}>
            </Table>
            <Button theme='light' type='primary' style={{marginRight: 8}} onClick={
                () => {
                    setEditingToken({
                        id: undefined,
                    });
                    setShowEdit(true);
                }
            }>添加令牌</Button>
            <Button label='复制所选令牌' type="warning" onClick={
                async () => {
                    if (selectedKeys.length === 0) {
                        showError('请至少选择一个令牌！');
                        return;
                    }
                    let keys = "";
                    for (let i = 0; i < selectedKeys.length; i++) {
                        keys += selectedKeys[i].name + "    sk-" + selectedKeys[i].key + "\n";
                    }
                    await copyText(keys);
                }
            }>复制所选令牌到剪贴板</Button>
        </>
    );
};

export default TokensTable;
