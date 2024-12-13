// ModelSettingsVisualEditor.js
import React, { useEffect, useState } from 'react';
import { Table, Button, Input, Modal, Form, Space } from '@douyinfe/semi-ui';
import { IconDelete, IconPlus, IconSearch } from '@douyinfe/semi-icons';
import { showError, showSuccess } from '../../../helpers';
import { API } from '../../../helpers';
export default function ModelSettingsVisualEditor(props) {
  const [models, setModels] = useState([]);
  const [visible, setVisible] = useState(false);
  const [currentModel, setCurrentModel] = useState(null);
  const [searchText, setSearchText] = useState('');
  const [currentPage, setCurrentPage] = useState(1);
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

  // 转换回JSON格式
  const generateJSONOutput = async () => {
    const output = {
      ModelPrice: {},
      ModelRatio: {},
      CompletionRatio: {}
    };
    let currentConvertModelName = '';
    try {
      models.forEach(model => {
        currentConvertModelName = model.name;
        if (model.price !== '') output.ModelPrice[model.name] = parseFloat(model.price);
        if (model.ratio !== '') output.ModelRatio[model.name] = parseFloat(model.ratio);
        if (model.completionRatio != '') output.CompletionRatio[model.name] = parseFloat(model.completionRatio);
      });
    } catch (error) {
      console.error('JSON转换错误:', error);
      showError('JSON转换错误, 请检查输入+模型名称: ' + currentConvertModelName);
      return;
    }

    const finalOutput = {
      ModelPrice: JSON.stringify(output.ModelPrice, null, 2),
      ModelRatio: JSON.stringify(output.ModelRatio, null, 2),
      CompletionRatio: JSON.stringify(output.CompletionRatio, null, 2)
    }

    forEach(finalOutput, (value, key) => {
      API.put('/api/option/', {
        key: key,
        value
      }).then(res => {
        if (res.data.success) {
          showSuccess('保存成功');
        } else {
          showError(res.data.message);
        }
      })
    })


    showSuccess('转换成功');
    props.refresh();
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
          placeholder="无"
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
          placeholder="默认倍率"
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
          placeholder="默认补全值"
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
    setModels(prev => [...prev, {
      name: values.name,
      price: values.price || '',
      ratio: values.ratio || '',
      completionRatio: values.completionRatio || ''
    }]);
    setVisible(false);
  };


  return (
    <>
      <Space vertical align="start" style={{ width: '100%' }}>
        <Space>
          <Button icon={<IconPlus />} onClick={() => setVisible(true)}>
            添加模型
          </Button>
          <Button type="primary" onClick={generateJSONOutput}>
            保存更改
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
          <Form.Input
            field="name"
            label="模型名称"
            required
            onChange={value => setCurrentModel(prev => ({ ...prev, name: value }))}
          />
          <Form.Input
            field="price"
            label="固定价格"
            onChange={value => setCurrentModel(prev => ({ ...prev, price: value }))}
          />
          <Form.Input
            field="ratio"
            label="模型倍率"
            onChange={value => setCurrentModel(prev => ({ ...prev, ratio: value }))}
          />
          <Form.Input
            field="completionRatio"
            label="补全倍率"
            onChange={value => setCurrentModel(prev => ({ ...prev, completionRatio: value }))}
          />
        </Form>
      </Modal>
    </>
  );
}
