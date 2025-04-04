import React, { useEffect, useState } from 'react';
import {
  Table,
  Button,
  Input,
  Modal,
  Form,
  Space,
  Typography,
  Radio,
  Notification,
} from '@douyinfe/semi-ui';
import {
  IconDelete,
  IconPlus,
  IconSearch,
  IconSave,
  IconBolt,
} from '@douyinfe/semi-icons';
import { showError, showSuccess } from '../../../helpers';
import { API } from '../../../helpers';
import { useTranslation } from 'react-i18next';

export default function ModelRatioNotSetEditor(props) {
  const { t } = useTranslation();
  const [models, setModels] = useState([]);
  const [visible, setVisible] = useState(false);
  const [batchVisible, setBatchVisible] = useState(false);
  const [currentModel, setCurrentModel] = useState(null);
  const [searchText, setSearchText] = useState('');
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [loading, setLoading] = useState(false);
  const [enabledModels, setEnabledModels] = useState([]);
  const [selectedRowKeys, setSelectedRowKeys] = useState([]);
  const [batchFillType, setBatchFillType] = useState('ratio');
  const [batchFillValue, setBatchFillValue] = useState('');
  const [batchRatioValue, setBatchRatioValue] = useState('');
  const [batchCompletionRatioValue, setBatchCompletionRatioValue] =
    useState('');
  const { Text } = Typography;
  // 定义可选的每页显示条数
  const pageSizeOptions = [10, 20, 50, 100];

  const getAllEnabledModels = async () => {
    try {
      const res = await API.get('/api/channel/models_enabled');
      const { success, message, data } = res.data;
      if (success) {
        setEnabledModels(data);
      } else {
        showError(message);
      }
    } catch (error) {
      console.error(t('获取启用模型失败:'), error);
      showError(t('获取启用模型失败'));
    }
  };

  useEffect(() => {
    // 获取所有启用的模型
    getAllEnabledModels();
  }, []);

  useEffect(() => {
    try {
      const modelPrice = JSON.parse(props.options.ModelPrice || '{}');
      const modelRatio = JSON.parse(props.options.ModelRatio || '{}');
      const completionRatio = JSON.parse(props.options.CompletionRatio || '{}');

      // 找出所有未设置价格和倍率的模型
      const unsetModels = enabledModels.filter((modelName) => {
        const hasPrice = modelPrice[modelName] !== undefined;
        const hasRatio = modelRatio[modelName] !== undefined;

        // 如果模型没有价格或者没有倍率设置，则显示
        return !hasPrice && !hasRatio;
      });

      // 创建模型数据
      const modelData = unsetModels.map((name) => ({
        name,
        price: modelPrice[name] || '',
        ratio: modelRatio[name] || '',
        completionRatio: completionRatio[name] || '',
      }));

      setModels(modelData);
      // 清空选择
      setSelectedRowKeys([]);
    } catch (error) {
      console.error(t('JSON解析错误:'), error);
    }
  }, [props.options, enabledModels]);

  // 首先声明分页相关的工具函数
  const getPagedData = (data, currentPage, pageSize) => {
    const start = (currentPage - 1) * pageSize;
    const end = start + pageSize;
    return data.slice(start, end);
  };

  // 处理页面大小变化
  const handlePageSizeChange = (size) => {
    setPageSize(size);
    // 重新计算当前页，避免数据丢失
    const totalPages = Math.ceil(filteredModels.length / size);
    if (currentPage > totalPages) {
      setCurrentPage(totalPages || 1);
    }
  };

  // 在 return 语句之前，先处理过滤和分页逻辑
  const filteredModels = models.filter((model) =>
    searchText
      ? model.name.toLowerCase().includes(searchText.toLowerCase())
      : true,
  );

  // 然后基于过滤后的数据计算分页数据
  const pagedData = getPagedData(filteredModels, currentPage, pageSize);

  const SubmitData = async () => {
    setLoading(true);
    const output = {
      ModelPrice: JSON.parse(props.options.ModelPrice || '{}'),
      ModelRatio: JSON.parse(props.options.ModelRatio || '{}'),
      CompletionRatio: JSON.parse(props.options.CompletionRatio || '{}'),
    };

    try {
      // 数据转换 - 只处理已修改的模型
      models.forEach((model) => {
        // 只有当用户设置了值时才更新
        if (model.price !== '') {
          // 如果价格不为空，则转换为浮点数，忽略倍率参数
          output.ModelPrice[model.name] = parseFloat(model.price);
        } else {
          if (model.ratio !== '')
            output.ModelRatio[model.name] = parseFloat(model.ratio);
          if (model.completionRatio !== '')
            output.CompletionRatio[model.name] = parseFloat(
              model.completionRatio,
            );
        }
      });

      // 准备API请求数组
      const finalOutput = {
        ModelPrice: JSON.stringify(output.ModelPrice, null, 2),
        ModelRatio: JSON.stringify(output.ModelRatio, null, 2),
        CompletionRatio: JSON.stringify(output.CompletionRatio, null, 2),
      };

      const requestQueue = Object.entries(finalOutput).map(([key, value]) => {
        return API.put('/api/option/', {
          key,
          value,
        });
      });

      // 批量处理请求
      const results = await Promise.all(requestQueue);

      // 验证结果
      if (requestQueue.length === 1) {
        if (results.includes(undefined)) return;
      } else if (requestQueue.length > 1) {
        if (results.includes(undefined)) {
          return showError(t('部分保存失败，请重试'));
        }
      }

      // 检查每个请求的结果
      for (const res of results) {
        if (!res.data.success) {
          return showError(res.data.message);
        }
      }

      showSuccess(t('保存成功'));
      props.refresh();
      // 重新获取未设置的模型
      getAllEnabledModels();
    } catch (error) {
      console.error(t('保存失败:'), error);
      showError(t('保存失败，请重试'));
    } finally {
      setLoading(false);
    }
  };

  const columns = [
    {
      title: t('模型名称'),
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: t('模型固定价格'),
      dataIndex: 'price',
      key: 'price',
      render: (text, record) => (
        <Input
          value={text}
          placeholder={t('按量计费')}
          onChange={(value) => updateModel(record.name, 'price', value)}
        />
      ),
    },
    {
      title: t('模型倍率'),
      dataIndex: 'ratio',
      key: 'ratio',
      render: (text, record) => (
        <Input
          value={text}
          placeholder={record.price !== '' ? t('模型倍率') : t('输入模型倍率')}
          disabled={record.price !== ''}
          onChange={(value) => updateModel(record.name, 'ratio', value)}
        />
      ),
    },
    {
      title: t('补全倍率'),
      dataIndex: 'completionRatio',
      key: 'completionRatio',
      render: (text, record) => (
        <Input
          value={text}
          placeholder={record.price !== '' ? t('补全倍率') : t('输入补全倍率')}
          disabled={record.price !== ''}
          onChange={(value) =>
            updateModel(record.name, 'completionRatio', value)
          }
        />
      ),
    },
  ];

  const updateModel = (name, field, value) => {
    if (value !== '' && isNaN(value)) {
      showError(t('请输入数字'));
      return;
    }
    setModels((prev) =>
      prev.map((model) =>
        model.name === name ? { ...model, [field]: value } : model,
      ),
    );
  };

  const addModel = (values) => {
    // 检查模型名称是否存在, 如果存在则拒绝添加
    if (models.some((model) => model.name === values.name)) {
      showError(t('模型名称已存在'));
      return;
    }
    setModels((prev) => [
      {
        name: values.name,
        price: values.price || '',
        ratio: values.ratio || '',
        completionRatio: values.completionRatio || '',
      },
      ...prev,
    ]);
    setVisible(false);
    showSuccess(t('添加成功'));
  };

  // 批量填充功能
  const handleBatchFill = () => {
    if (selectedRowKeys.length === 0) {
      showError(t('请先选择需要批量设置的模型'));
      return;
    }

    if (batchFillType === 'bothRatio') {
      if (batchRatioValue === '' || batchCompletionRatioValue === '') {
        showError(t('请输入模型倍率和补全倍率'));
        return;
      }
      if (isNaN(batchRatioValue) || isNaN(batchCompletionRatioValue)) {
        showError(t('请输入有效的数字'));
        return;
      }
    } else {
      if (batchFillValue === '') {
        showError(t('请输入填充值'));
        return;
      }
      if (isNaN(batchFillValue)) {
        showError(t('请输入有效的数字'));
        return;
      }
    }

    // 根据选择的类型批量更新模型
    setModels((prev) =>
      prev.map((model) => {
        if (selectedRowKeys.includes(model.name)) {
          if (batchFillType === 'price') {
            return {
              ...model,
              price: batchFillValue,
              ratio: '',
              completionRatio: '',
            };
          } else if (batchFillType === 'ratio') {
            return {
              ...model,
              price: '',
              ratio: batchFillValue,
            };
          } else if (batchFillType === 'completionRatio') {
            return {
              ...model,
              price: '',
              completionRatio: batchFillValue,
            };
          } else if (batchFillType === 'bothRatio') {
            return {
              ...model,
              price: '',
              ratio: batchRatioValue,
              completionRatio: batchCompletionRatioValue,
            };
          }
        }
        return model;
      }),
    );

    setBatchVisible(false);
    Notification.success({
      title: t('批量设置成功'),
      content: t('已为 {{count}} 个模型设置{{type}}', {
        count: selectedRowKeys.length,
        type:
          batchFillType === 'price'
            ? t('固定价格')
            : batchFillType === 'ratio'
              ? t('模型倍率')
              : batchFillType === 'completionRatio'
                ? t('补全倍率')
                : t('模型倍率和补全倍率'),
      }),
      duration: 3,
    });
  };

  const handleBatchTypeChange = (value) => {
    console.log(t('Changing batch type to:'), value);
    setBatchFillType(value);

    // 切换类型时清空对应的值
    if (value !== 'bothRatio') {
      setBatchFillValue('');
    } else {
      setBatchRatioValue('');
      setBatchCompletionRatioValue('');
    }
  };

  const rowSelection = {
    selectedRowKeys,
    onChange: (selectedKeys) => {
      setSelectedRowKeys(selectedKeys);
    },
  };

  return (
    <>
      <Space vertical align='start' style={{ width: '100%' }}>
        <Space>
          <Button icon={<IconPlus />} onClick={() => setVisible(true)}>
            {t('添加模型')}
          </Button>
          <Button
            icon={<IconBolt />}
            type='secondary'
            onClick={() => setBatchVisible(true)}
            disabled={selectedRowKeys.length === 0}
          >
            {t('批量设置')} ({selectedRowKeys.length})
          </Button>
          <Button
            type='primary'
            icon={<IconSave />}
            onClick={SubmitData}
            loading={loading}
          >
            {t('应用更改')}
          </Button>
          <Input
            prefix={<IconSearch />}
            placeholder={t('搜索模型名称')}
            value={searchText}
            onChange={(value) => {
              setSearchText(value);
              setCurrentPage(1);
            }}
            style={{ width: 200 }}
          />
        </Space>

        <Text>
          {t('此页面仅显示未设置价格或倍率的模型，设置后将自动从列表中移除')}
        </Text>

        <Table
          columns={columns}
          dataSource={pagedData}
          rowSelection={rowSelection}
          rowKey='name'
          pagination={{
            currentPage: currentPage,
            pageSize: pageSize,
            total: filteredModels.length,
            onPageChange: (page) => setCurrentPage(page),
            onPageSizeChange: handlePageSizeChange,
            pageSizeOptions: pageSizeOptions,
            formatPageText: (page) =>
              t('第 {{start}} - {{end}} 条，共 {{total}} 条', {
                start: page.currentStart,
                end: page.currentEnd,
                total: filteredModels.length,
              }),
            showTotal: true,
            showSizeChanger: true,
          }}
          empty={
            <div style={{ textAlign: 'center', padding: '20px' }}>
              {t('没有未设置的模型')}
            </div>
          }
        />
      </Space>

      {/* 添加模型弹窗 */}
      <Modal
        title={t('添加模型')}
        visible={visible}
        onCancel={() => setVisible(false)}
        onOk={() => {
          currentModel && addModel(currentModel);
        }}
      >
        <Form>
          <Form.Input
            field='name'
            label={t('模型名称')}
            placeholder='strawberry'
            required
            onChange={(value) =>
              setCurrentModel((prev) => ({ ...prev, name: value }))
            }
          />
          <Form.Switch
            field='priceMode'
            label={
              <>
                {t('定价模式')}：
                {currentModel?.priceMode ? t('固定价格') : t('倍率模式')}
              </>
            }
            onChange={(checked) => {
              setCurrentModel((prev) => ({
                ...prev,
                price: '',
                ratio: '',
                completionRatio: '',
                priceMode: checked,
              }));
            }}
          />
          {currentModel?.priceMode ? (
            <Form.Input
              field='price'
              label={t('固定价格(每次)')}
              placeholder={t('输入每次价格')}
              onChange={(value) =>
                setCurrentModel((prev) => ({ ...prev, price: value }))
              }
            />
          ) : (
            <>
              <Form.Input
                field='ratio'
                label={t('模型倍率')}
                placeholder={t('输入模型倍率')}
                onChange={(value) =>
                  setCurrentModel((prev) => ({ ...prev, ratio: value }))
                }
              />
              <Form.Input
                field='completionRatio'
                label={t('补全倍率')}
                placeholder={t('输入补全价格')}
                onChange={(value) =>
                  setCurrentModel((prev) => ({
                    ...prev,
                    completionRatio: value,
                  }))
                }
              />
            </>
          )}
        </Form>
      </Modal>

      {/* 批量设置弹窗 */}
      <Modal
        title={t('批量设置模型参数')}
        visible={batchVisible}
        onCancel={() => setBatchVisible(false)}
        onOk={handleBatchFill}
        width={500}
      >
        <Form>
          <Form.Section text={t('设置类型')}>
            <div style={{ marginBottom: '16px' }}>
              <Space>
                <Radio
                  checked={batchFillType === 'price'}
                  onChange={() => handleBatchTypeChange('price')}
                >
                  {t('固定价格')}
                </Radio>
                <Radio
                  checked={batchFillType === 'ratio'}
                  onChange={() => handleBatchTypeChange('ratio')}
                >
                  {t('模型倍率')}
                </Radio>
                <Radio
                  checked={batchFillType === 'completionRatio'}
                  onChange={() => handleBatchTypeChange('completionRatio')}
                >
                  {t('补全倍率')}
                </Radio>
                <Radio
                  checked={batchFillType === 'bothRatio'}
                  onChange={() => handleBatchTypeChange('bothRatio')}
                >
                  {t('模型倍率和补全倍率同时设置')}
                </Radio>
              </Space>
            </div>
          </Form.Section>

          {batchFillType === 'bothRatio' ? (
            <>
              <Form.Input
                field='batchRatioValue'
                label={t('模型倍率值')}
                placeholder={t('请输入模型倍率')}
                value={batchRatioValue}
                onChange={(value) => setBatchRatioValue(value)}
              />
              <Form.Input
                field='batchCompletionRatioValue'
                label={t('补全倍率值')}
                placeholder={t('请输入补全倍率')}
                value={batchCompletionRatioValue}
                onChange={(value) => setBatchCompletionRatioValue(value)}
              />
            </>
          ) : (
            <Form.Input
              field='batchFillValue'
              label={
                batchFillType === 'price'
                  ? t('固定价格值')
                  : batchFillType === 'ratio'
                    ? t('模型倍率值')
                    : t('补全倍率值')
              }
              placeholder={t('请输入数值')}
              value={batchFillValue}
              onChange={(value) => setBatchFillValue(value)}
            />
          )}

          <Text type='tertiary'>
            {t('将为选中的 ')} <Text strong>{selectedRowKeys.length}</Text>{' '}
            {t(' 个模型设置相同的值')}
          </Text>
          <div style={{ marginTop: '8px' }}>
            <Text type='tertiary'>
              {t('当前设置类型: ')}{' '}
              <Text strong>
                {batchFillType === 'price'
                  ? t('固定价格')
                  : batchFillType === 'ratio'
                    ? t('模型倍率')
                    : batchFillType === 'completionRatio'
                      ? t('补全倍率')
                      : t('模型倍率和补全倍率')}
              </Text>
            </Text>
          </div>
        </Form>
      </Modal>
    </>
  );
}
