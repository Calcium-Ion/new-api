import React, { useContext, useEffect, useRef, useMemo, useState } from 'react';
import { API, copy, showError, showInfo, showSuccess } from '../helpers';

import {
  Banner,
  Input,
  Layout,
  Modal,
  Space,
  Table,
  Tag,
  Tooltip,
  Popover,
  ImagePreview,
  Button,
} from '@douyinfe/semi-ui';
import {
  IconMore,
  IconVerify,
  IconUploadError,
  IconHelpCircle,
} from '@douyinfe/semi-icons';
import { UserContext } from '../context/User/index.js';
import Text from '@douyinfe/semi-ui/lib/es/typography/text';

function renderQuotaType(type) {
  // Ensure all cases are string literals by adding quotes.
  switch (type) {
    case 1:
      return (
        <Tag color='teal' size='large'>
          按次计费
        </Tag>
      );
    case 0:
      return (
        <Tag color='violet' size='large'>
          按量计费
        </Tag>
      );
    default:
      return '未知';
  }
}

function renderAvailable(available) {
  return available ? (
    <Popover
        content={
          <div style={{ padding: 8 }}>您的分组可以使用该模型</div>
        }
        position='top'
        key={available}
        style={{
            backgroundColor: 'rgba(var(--semi-blue-4),1)',
            borderColor: 'rgba(var(--semi-blue-4),1)',
            color: 'var(--semi-color-white)',
            borderWidth: 1,
            borderStyle: 'solid',
        }}
    >
        <IconVerify style={{ color: 'green' }}  size="large" />
    </Popover>
  ) : (
    <Popover
        content={
          <div style={{ padding: 8 }}>您的分组无权使用该模型</div>
        }
        position='top'
        key={available}
        style={{
            backgroundColor: 'rgba(var(--semi-blue-4),1)',
            borderColor: 'rgba(var(--semi-blue-4),1)',
            color: 'var(--semi-color-white)',
            borderWidth: 1,
            borderStyle: 'solid',
        }}
    >
        <IconUploadError style={{ color: '#FFA54F' }}  size="large" />
    </Popover>
  );
}

const ModelPricing = () => {
  const [filteredValue, setFilteredValue] = useState([]);
  const compositionRef = useRef({ isComposition: false });
  const [selectedRowKeys, setSelectedRowKeys] = useState([]);
  const [modalImageUrl, setModalImageUrl] = useState('');
  const [isModalOpenurl, setIsModalOpenurl] = useState(false);
  const [selectedGroup, setSelectedGroup] = useState('default');

  const rowSelection = useMemo(
      () => ({
          onChange: (selectedRowKeys, selectedRows) => {
            setSelectedRowKeys(selectedRowKeys);
          },
      }),
      []
  );

  const handleChange = (value) => {
    if (compositionRef.current.isComposition) {
      return;
    }
    const newFilteredValue = value ? [value] : [];
    setFilteredValue(newFilteredValue);
  };
  const handleCompositionStart = () => {
    compositionRef.current.isComposition = true;
  };

  const handleCompositionEnd = (event) => {
    compositionRef.current.isComposition = false;
    const value = event.target.value;
    const newFilteredValue = value ? [value] : [];
    setFilteredValue(newFilteredValue);
  };

  const columns = [
    {
      title: '可用性',
      dataIndex: 'available',
      render: (text, record, index) => {
         // if record.enable_groups contains selectedGroup, then available is true
        return renderAvailable(record.enable_groups.includes(selectedGroup));
      },
      sorter: (a, b) => a.available - b.available,
    },
    {
      title: (
        <Space>
          <span>模型名称</span>
          <Input
            placeholder='模糊搜索'
            style={{ width: 200 }}
            onCompositionStart={handleCompositionStart}
            onCompositionEnd={handleCompositionEnd}
            onChange={handleChange}
            showClear
          />
        </Space>
      ),
      dataIndex: 'model_name', // 以finish_time作为dataIndex
      render: (text, record, index) => {
        return (
          <>
            <Tag
              color='green'
              size='large'
              onClick={() => {
                copyText(text);
              }}
            >
              {text}
            </Tag>
          </>
        );
      },
      onFilter: (value, record) =>
        record.model_name.toLowerCase().includes(value.toLowerCase()),
      filteredValue,
    },
    {
      title: '计费类型',
      dataIndex: 'quota_type',
      render: (text, record, index) => {
        return renderQuotaType(parseInt(text));
      },
      sorter: (a, b) => a.quota_type - b.quota_type,
    },
    {
      title: '可用分组',
      dataIndex: 'enable_groups',
      render: (text, record, index) => {
        // enable_groups is a string array
        return (
          <Space>
            {text.map((group) => {
              if (group === selectedGroup) {
                return (
                  <Tag
                    color='blue'
                    size='large'
                    prefixIcon={<IconVerify />}
                  >
                    {group}
                  </Tag>
                );
              } else {
                return (
                  <Tag
                    color='blue'
                    size='large'
                    onClick={() => {
                      setSelectedGroup(group);
                      showInfo('当前查看的分组为：' + group + '，倍率为：' + groupRatio[group]);
                    }}
                  >
                    {group}
                  </Tag>
                );
              }
            })}
          </Space>
        );
      },
    },
    {
      title: () => (
        <span style={{'display':'flex','alignItems':'center'}}>
          倍率
          <Popover
            content={
              <div style={{ padding: 8 }}>倍率是为了方便换算不同价格的模型<br/>点击查看倍率说明</div>
            }
            position='top'
            style={{
                backgroundColor: 'rgba(var(--semi-blue-4),1)',
                borderColor: 'rgba(var(--semi-blue-4),1)',
                color: 'var(--semi-color-white)',
                borderWidth: 1,
                borderStyle: 'solid',
            }}
          >
            <IconHelpCircle
              onClick={() => {
                setModalImageUrl('/ratio.png');
                setIsModalOpenurl(true);
              }}
            />
          </Popover>
        </span>
      ),
      dataIndex: 'model_ratio',
      render: (text, record, index) => {
        let content = text;
        let completionRatio = parseFloat(record.completion_ratio.toFixed(3));
        content = (
          <>
            <Text>模型：{record.quota_type === 0 ? text : '无'}</Text>
            <br />
            <Text>补全：{record.quota_type === 0 ? completionRatio : '无'}</Text>
            <br />
            <Text>分组：{groupRatio[selectedGroup]}</Text>
          </>
        );
        return <div>{content}</div>;
      },
    },
    {
      title: '模型价格',
      dataIndex: 'model_price',
      render: (text, record, index) => {
        let content = text;
        if (record.quota_type === 0) {
          // 这里的 *2 是因为 1倍率=0.002刀，请勿删除
          let inputRatioPrice = record.model_ratio * 2 * groupRatio[selectedGroup];
          let completionRatioPrice =
            record.model_ratio *
            record.completion_ratio * 2 *
            groupRatio[selectedGroup];
          content = (
            <>
              <Text>提示 ${inputRatioPrice} / 1M tokens</Text>
              <br />
              <Text>补全 ${completionRatioPrice} / 1M tokens</Text>
            </>
          );
        } else {
          let price = parseFloat(text) * groupRatio[selectedGroup];
          content = <>模型价格：${price}</>;
        }
        return <div>{content}</div>;
      },
    },
  ];

  const [models, setModels] = useState([]);
  const [loading, setLoading] = useState(true);
  const [userState, userDispatch] = useContext(UserContext);
  const [groupRatio, setGroupRatio] = useState({});

  const setModelsFormat = (models, groupRatio) => {
    for (let i = 0; i < models.length; i++) {
      models[i].key = models[i].model_name;
      models[i].group_ratio = groupRatio[models[i].model_name];
    }
    // sort by quota_type
    models.sort((a, b) => {
      return a.quota_type - b.quota_type;
    });

    // sort by model_name, start with gpt is max, other use localeCompare
    models.sort((a, b) => {
      if (a.model_name.startsWith('gpt') && !b.model_name.startsWith('gpt')) {
        return -1;
      } else if (
        !a.model_name.startsWith('gpt') &&
        b.model_name.startsWith('gpt')
      ) {
        return 1;
      } else {
        return a.model_name.localeCompare(b.model_name);
      }
    });

    setModels(models);
  };

  const loadPricing = async () => {
    setLoading(true);

    let url = '';
    url = `/api/pricing`;
    const res = await API.get(url);
    const { success, message, data, group_ratio } = res.data;
    if (success) {
      setGroupRatio(group_ratio);
      setSelectedGroup(userState.user ? userState.user.group : 'default')
      setModelsFormat(data, group_ratio);
    } else {
      showError(message);
    }
    setLoading(false);
  };

  const refresh = async () => {
    await loadPricing();
  };

  const copyText = async (text) => {
    if (await copy(text)) {
      showSuccess('已复制：' + text);
    } else {
      // setSearchKeyword(text);
      Modal.error({ title: '无法复制到剪贴板，请手动复制', content: text });
    }
  };

  useEffect(() => {
    refresh().then();
  }, []);

  return (
    <>
      <Layout>
        {userState.user ? (
          <Banner
            type="success"
            fullMode={false}
            closeIcon="null"
            description={`您的默认分组为：${userState.user.group}，分组倍率为：${groupRatio[userState.user.group]}`}
          />
        ) : (
          <Banner
            type='warning'
            fullMode={false}
            closeIcon="null"
            description={`您还未登陆，显示的价格为默认分组倍率: ${groupRatio['default']}`}
          />
        )}
        <br/>
        <Banner 
            type="info"
            fullMode={false}
            description={<div>按量计费费用 = 分组倍率 × 模型倍率 × （提示token数 + 补全token数 × 补全倍率）/ 500000 （单位：美元）</div>}
            closeIcon="null"
        />
        <br/>
        <Button
          theme='light'
          type='tertiary'
          style={{width: 150}}
          onClick={() => {
            copyText(selectedRowKeys);
          }}
          disabled={selectedRowKeys == ""}
        >
          复制选中模型
        </Button>
        <Table
          style={{ marginTop: 5 }}
          columns={columns}
          dataSource={models}
          loading={loading}
          pagination={{
            pageSize: models.length,
            showSizeChanger: false,
          }}
          rowSelection={rowSelection}
        />
        <ImagePreview
          src={modalImageUrl}
          visible={isModalOpenurl}
          onVisibleChange={(visible) => setIsModalOpenurl(visible)}
        />
      </Layout>
    </>
  );
};

export default ModelPricing;
