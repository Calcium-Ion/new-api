// ModelSettingsVisualEditor.js
import React, { useEffect, useState } from 'react';
import { Table, Button, Input, Modal, Form, Space } from '@douyinfe/semi-ui';
import { IconDelete, IconPlus, IconSearch, IconSave } from '@douyinfe/semi-icons';
import { showError, showSuccess } from '../../../helpers';
import { API } from '../../../helpers';
export default function ModelSettingsVisualEditor(props) {
  const [models, setModels] = useState([]);
  const [visible, setVisible] = useState(false);
  const [currentModel, setCurrentModel] = useState(null);
  const [searchText, setSearchText] = useState('');
  const [currentPage, setCurrentPage] = useState(1);
  const [loading, setLoading] = useState(false);
  const pageSize = 10;

  useEffect(() => {
    try {
      const modelPrice = JSON.parse(props.options.ModelPrice || '{}');
      const modelRatio = JSON.parse(props.options.ModelRatio || '{}');
      const completionRatio = JSON.parse(props.options.CompletionRatio || '{}');

      // 合并所有模型名称
      const modelNames = new Set([
        ...Object.keys(modelPrice),
        ...Object.keys(modelRatio),
        ...Object.keys(completionRatio)
      ]);

      const modelData = Array.from(modelNames).map(name => ({
        name,
        price: modelPrice[name] === undefined ? '' : modelPrice[name],
        ratio: modelRatio[name] === undefined ? '' : modelRatio[name],
        completionRatio: completionRatio[name] === undefined ? '' : completionRatio[name]
      }));

      setModels(modelData);
    } catch (error) {
      console.error('JSON解析错误:', error);
    }
  }, [props.options]);

  // 首先声明分页相关的工具函数
  const getPagedData = (data, currentPage, pageSize) => {
    const start = (currentPage - 1) * pageSize;
    const end = start + pageSize;
    return data.slice(start, end);
  };

  // 在 return 语句之前，先处理过滤和分页逻辑
  const filteredModels = models.filter(model =>
    searchText ? model.name.toLowerCase().includes(searchText.toLowerCase()) : true
  );

  // 然后基于过滤后的数据计算分页数据
  const pagedData = getPagedData(filteredModels, currentPage, pageSize);

  const SubmitData = async () => {
    setLoading(true);
    const output = {
      ModelPrice: {},
      ModelRatio: {},
      CompletionRatio: {}
    };
    let currentConvertModelName = '';

    try {
      // 数据转换
      models.forEach(model => {
        currentConvertModelName = model.name;
        if (model.price !== '') {
          // 如果价格不为空，则转换为浮点数，忽略倍率参数
          output.ModelPrice[model.name] = parseFloat(model.price)
        } else {
          if (model.ratio !== '') output.ModelRatio[model.name] = parseFloat(model.ratio);
          if (model.completionRatio != '') output.CompletionRatio[model.name] = parseFloat(model.completionRatio);
        }
      });

      // 准备API请求数组
      const finalOutput = {
        ModelPrice: JSON.stringify(output.ModelPrice, null, 2),
        ModelRatio: JSON.stringify(output.ModelRatio, null, 2),
        CompletionRatio: JSON.stringify(output.CompletionRatio, null, 2)
      };

      const requestQueue = Object.entries(finalOutput).map(([key, value]) => {
        return API.put('/api/option/', {
          key,
          value
        });
      });

      // 批量处理请求
      const results = await Promise.all(requestQueue);

      // 验证结果
      if (requestQueue.length === 1) {
        if (results.includes(undefined)) return;
      } else if (requestQueue.length > 1) {
        if (results.includes(undefined)) {
          return showError('部分保存失败，请重试');
        }
      }

      // 检查每个请求的结果
      for (const res of results) {
        if (!res.data.success) {
          return showError(res.data.message);
        }
      }

      showSuccess('保存成功');
      props.refresh();

    } catch (error) {
      console.error('保存失败:', error);
      showError('保存失败，请重试');
    } finally {
      setLoading(false);
    }
  };

  const columns = [
    {
      title: '模型名称',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: '固定价格',
      dataIndex: 'price',
      key: 'price',
      render: (text, record) => (
        <Input
          value={text}
          placeholder="按量计价"
          onChange={value => updateModel(record.name, 'price', value)}
        />
      )
    },
    {
      title: '模型倍率',
      dataIndex: 'ratio',
      key: 'ratio',
      render: (text, record) => (
        <Input
          value={text}

          placeholder={record.price !== '' ? '固定价格' : '默认补全倍率'}
          disabled={record.price !== ''}
          onChange={value => updateModel(record.name, 'ratio', value)}
        />
      )
    },
    {
      title: '补全倍率',
      dataIndex: 'completionRatio',
      key: 'completionRatio',
      render: (text, record) => (
        <Input
          value={text}
          placeholder={record.price !== '' ? '固定价格' : '默认补全倍率'}
          disabled={record.price !== ''}
          onChange={value => updateModel(record.name, 'completionRatio', value)}
        />
      )
    },
    {
      title: '操作',
      key: 'action',
      render: (_, record) => (
        <Button
          icon={<IconDelete />}
          type="danger"
          onClick={() => deleteModel(record.name)}
        />
      )
    }
  ];

  const updateModel = (name, field, value) => {
    if (isNaN(value)) {
      showError('请输入数字');
      return;
    }
    setModels(prev =>
      prev.map(model =>
        model.name === name
          ? { ...model, [field]: value }
          : model
      )
    );
  };

  const deleteModel = (name) => {
    setModels(prev => prev.filter(model => model.name !== name));
  };
  const addModel = (values) => {
    // 检查模型名称是否存在, 如果存在则拒绝添加
    if (models.some(model => model.name === values.name)) {
      showError('模型名称已存在');
      return;
    }
    // 不允许同时添加固定价格和倍率
    if (values.price !== '' && (values.ratio !== '' || values.completionRatio !== '')) {
      showError('固定价格和倍率不能同时存在');
      return;
    }
    setModels(prev => [{
      name: values.name,
      price: values.price || '',
      ratio: values.ratio || '',
      completionRatio: values.completionRatio || ''
    }, ...prev]);
    setVisible(false);
    showSuccess('添加成功');
  };


  return (
    <>
      <h3>模型价格</h3>
      <Space vertical align="start" style={{ width: '100%' }}>
        <Space>
          <Button icon={<IconPlus />} onClick={() => setVisible(true)}>
            添加模型
          </Button>
          <Button type="primary" icon={<IconSave />} onClick={SubmitData}>
            应用更改
          </Button>
          <Input
            prefix={<IconSearch />}
            placeholder="搜索模型名称"
            value={searchText}
            onChange={value => {
              setSearchText(value)
              // 搜索时重置页码
              setCurrentPage(1);
            }}
            style={{ width: 200 }}
          />
        </Space>
        <Table
          columns={columns}
          dataSource={pagedData} // 使用分页后的数据
          pagination={{
            currentPage: currentPage,
            pageSize: pageSize,
            total: filteredModels.length,
            onPageChange: page => setCurrentPage(page),
            showTotal: true,
            showSizeChanger: false
          }}
        />
      </Space>

      <Modal
        title="添加模型"
        visible={visible}
        onCancel={() => setVisible(false)}
        onOk={() => {
          currentModel && addModel(currentModel);
        }}
      >
        <Form>
          <p>请输入固定价格或者模型倍率+补全倍率</p>
          <Form.Input
            field="name"
            label="模型名称"
            placeholder="strawberry"
            required
            onChange={value => setCurrentModel(prev => ({ ...prev, name: value }))}
          />
          <Form.Input
            field="price"
            label="固定价格(每次)"
            placeholder="输入每次价格"
            onChange={value => setCurrentModel(prev => ({ ...prev, price: value }))}
          />
          <Form.Input
            field="ratio"
            label="模型倍率"
            placeholder="输入模型倍率"
            onChange={value => setCurrentModel(prev => ({ ...prev, ratio: value }))}
          />
          <Form.Input
            field="completionRatio"
            label="补全倍率"
            placeholder="输入补全价格"
            onChange={value => setCurrentModel(prev => ({ ...prev, completionRatio: value }))}
          />
        </Form>
      </Modal>
    </>
  );
}
