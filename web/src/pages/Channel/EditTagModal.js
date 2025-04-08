import React, { useState, useEffect } from 'react';
import {
  API,
  showError,
  showInfo,
  showSuccess,
  showWarning,
  verifyJSON,
} from '../../helpers';
import {
  SideSheet,
  Space,
  Button,
  Input,
  Typography,
  Spin,
  Modal,
  Select,
  Banner,
  TextArea,
} from '@douyinfe/semi-ui';
import TextInput from '../../components/custom/TextInput.js';
import { getChannelModels } from '../../components/utils.js';

const MODEL_MAPPING_EXAMPLE = {
  'gpt-3.5-turbo': 'gpt-3.5-turbo-0125',
};

const EditTagModal = (props) => {
  const { visible, tag, handleClose, refresh } = props;
  const [loading, setLoading] = useState(false);
  const [originModelOptions, setOriginModelOptions] = useState([]);
  const [modelOptions, setModelOptions] = useState([]);
  const [groupOptions, setGroupOptions] = useState([]);
  const [basicModels, setBasicModels] = useState([]);
  const [fullModels, setFullModels] = useState([]);
  const [customModel, setCustomModel] = useState('');
  const originInputs = {
    tag: '',
    new_tag: null,
    model_mapping: null,
    groups: [],
    models: [],
  };
  const [inputs, setInputs] = useState(originInputs);

  const handleInputChange = (name, value) => {
    setInputs((inputs) => ({ ...inputs, [name]: value }));
    if (name === 'type') {
      let localModels = [];
      switch (value) {
        case 2:
          localModels = [
            'mj_imagine',
            'mj_variation',
            'mj_reroll',
            'mj_blend',
            'mj_upscale',
            'mj_describe',
            'mj_uploads',
          ];
          break;
        case 5:
          localModels = [
            'swap_face',
            'mj_imagine',
            'mj_variation',
            'mj_reroll',
            'mj_blend',
            'mj_upscale',
            'mj_describe',
            'mj_zoom',
            'mj_shorten',
            'mj_modal',
            'mj_inpaint',
            'mj_custom_zoom',
            'mj_high_variation',
            'mj_low_variation',
            'mj_pan',
            'mj_uploads',
          ];
          break;
        case 36:
          localModels = ['suno_music', 'suno_lyrics'];
          break;
        default:
          localModels = getChannelModels(value);
          break;
      }
      if (inputs.models.length === 0) {
        setInputs((inputs) => ({ ...inputs, models: localModels }));
      }
      setBasicModels(localModels);
    }
  };

  const fetchModels = async () => {
    try {
      let res = await API.get(`/api/channel/models`);
      let localModelOptions = res.data.data.map((model) => ({
        label: model.id,
        value: model.id,
      }));
      setOriginModelOptions(localModelOptions);
      setFullModels(res.data.data.map((model) => model.id));
      setBasicModels(
        res.data.data
          .filter((model) => {
            return model.id.startsWith('gpt-') || model.id.startsWith('text-');
          })
          .map((model) => model.id),
      );
    } catch (error) {
      showError(error.message);
    }
  };

  const fetchGroups = async () => {
    try {
      let res = await API.get(`/api/group/`);
      if (res === undefined) {
        return;
      }
      setGroupOptions(
        res.data.data.map((group) => ({
          label: group,
          value: group,
        })),
      );
    } catch (error) {
      showError(error.message);
    }
  };

  const handleSave = async () => {
    setLoading(true);
    let data = {
      tag: tag,
    };
    if (inputs.model_mapping !== null && inputs.model_mapping !== '') {
      if (inputs.model_mapping !== '' && !verifyJSON(inputs.model_mapping)) {
        showInfo('模型映射必须是合法的 JSON 格式！');
        setLoading(false);
        return;
      }
      data.model_mapping = inputs.model_mapping;
    }
    if (inputs.groups.length > 0) {
      data.groups = inputs.groups.join(',');
    }
    if (inputs.models.length > 0) {
      data.models = inputs.models.join(',');
    }
    data.new_tag = inputs.new_tag;
    // check have any change
    if (
      data.model_mapping === undefined &&
      data.groups === undefined &&
      data.models === undefined &&
      data.new_tag === undefined
    ) {
      showWarning('没有任何修改！');
      setLoading(false);
      return;
    }
    await submit(data);
    setLoading(false);
  };

  const submit = async (data) => {
    try {
      const res = await API.put('/api/channel/tag', data);
      if (res?.data?.success) {
        showSuccess('标签更新成功！');
        refresh();
        handleClose();
      }
    } catch (error) {
      showError(error);
    }
  };

  useEffect(() => {
    let localModelOptions = [...originModelOptions];
    inputs.models.forEach((model) => {
      if (!localModelOptions.find((option) => option.label === model)) {
        localModelOptions.push({
          label: model,
          value: model,
        });
      }
    });
    setModelOptions(localModelOptions);
  }, [originModelOptions, inputs.models]);

  useEffect(() => {
    setInputs({
      ...originInputs,
      tag: tag,
      new_tag: tag,
    });
    fetchModels().then();
    fetchGroups().then();
  }, [visible]);

  const addCustomModels = () => {
    if (customModel.trim() === '') return;
    // 使用逗号分隔字符串，然后去除每个模型名称前后的空格
    const modelArray = customModel.split(',').map((model) => model.trim());

    let localModels = [...inputs.models];
    let localModelOptions = [...modelOptions];
    let hasError = false;

    modelArray.forEach((model) => {
      // 检查模型是否已存在，且模型名称非空
      if (model && !localModels.includes(model)) {
        localModels.push(model); // 添加到模型列表
        localModelOptions.push({
          // 添加到下拉选项
          key: model,
          text: model,
          value: model,
        });
      } else if (model) {
        showError('某些模型已存在！');
        hasError = true;
      }
    });

    if (hasError) return; // 如果有错误则终止操作

    // 更新状态值
    setModelOptions(localModelOptions);
    setCustomModel('');
    handleInputChange('models', localModels);
  };

  return (
    <SideSheet
      title='编辑标签'
      visible={visible}
      onCancel={handleClose}
      footer={
        <div style={{ display: 'flex', justifyContent: 'flex-end' }}>
          <Space>
            <Button onClick={handleClose}>取消</Button>
            <Button type='primary' onClick={handleSave} loading={loading}>
              保存
            </Button>
          </Space>
        </div>
      }
    >
      <div style={{ marginTop: 10 }}>
        <Banner
          type={'warning'}
          description={<>所有编辑均为覆盖操作，留空则不更改</>}
        ></Banner>
      </div>
      <Spin spinning={loading}>
        <TextInput
          label='标签名，留空则解散标签'
          name='newTag'
          value={inputs.new_tag}
          onChange={(value) => setInputs({ ...inputs, new_tag: value })}
          placeholder='请输入新标签'
        />
        <div style={{ marginTop: 10 }}>
          <Typography.Text strong>模型，留空则不更改：</Typography.Text>
        </div>
        <Select
          placeholder={'请选择该渠道所支持的模型，留空则不更改'}
          name='models'
          required
          multiple
          selection
          filter
          searchPosition='dropdown'
          onChange={(value) => {
            handleInputChange('models', value);
          }}
          value={inputs.models}
          autoComplete='new-password'
          optionList={modelOptions}
        />
        <Input
          addonAfter={
            <Button type='primary' onClick={addCustomModels}>
              填入
            </Button>
          }
          placeholder='输入自定义模型名称'
          value={customModel}
          onChange={(value) => {
            setCustomModel(value.trim());
          }}
        />
        <div style={{ marginTop: 10 }}>
          <Typography.Text strong>分组，留空则不更改：</Typography.Text>
        </div>
        <Select
          placeholder={'请选择可以使用该渠道的分组，留空则不更改'}
          name='groups'
          required
          multiple
          selection
          allowAdditions
          additionLabel={'请在系统设置页面编辑分组倍率以添加新的分组：'}
          onChange={(value) => {
            handleInputChange('groups', value);
          }}
          value={inputs.groups}
          autoComplete='new-password'
          optionList={groupOptions}
        />
        <div style={{ marginTop: 10 }}>
          <Typography.Text strong>模型重定向：</Typography.Text>
        </div>
        <TextArea
          placeholder={`此项可选，用于修改请求体中的模型名称，为一个 JSON 字符串，键为请求中模型名称，值为要替换的模型名称，留空则不更改`}
          name='model_mapping'
          onChange={(value) => {
            handleInputChange('model_mapping', value);
          }}
          autosize
          value={inputs.model_mapping}
          autoComplete='new-password'
        />
        <Space>
          <Typography.Text
            style={{
              color: 'rgba(var(--semi-blue-5), 1)',
              userSelect: 'none',
              cursor: 'pointer',
            }}
            onClick={() => {
              handleInputChange(
                'model_mapping',
                JSON.stringify(MODEL_MAPPING_EXAMPLE, null, 2),
              );
            }}
          >
            填入模板
          </Typography.Text>
          <Typography.Text
            style={{
              color: 'rgba(var(--semi-blue-5), 1)',
              userSelect: 'none',
              cursor: 'pointer',
            }}
            onClick={() => {
              handleInputChange('model_mapping', JSON.stringify({}, null, 2));
            }}
          >
            清空重定向
          </Typography.Text>
          <Typography.Text
            style={{
              color: 'rgba(var(--semi-blue-5), 1)',
              userSelect: 'none',
              cursor: 'pointer',
            }}
            onClick={() => {
              handleInputChange('model_mapping', '');
            }}
          >
            不更改
          </Typography.Text>
        </Space>
      </Spin>
    </SideSheet>
  );
};

export default EditTagModal;
