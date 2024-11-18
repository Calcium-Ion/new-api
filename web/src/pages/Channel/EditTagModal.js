import React, { useState, useEffect } from 'react';
import { API, showError, showSuccess } from '../../helpers';
import { SideSheet, Space, Button, Input, Typography, Spin, Modal } from '@douyinfe/semi-ui';
import TextInput from '../../components/TextInput.js';

const EditTagModal = (props) => {
  const { visible, tag, handleClose, refresh } = props;
  const [loading, setLoading] = useState(false);
  const originInputs = {
    tag: '',
    new_tag: null,
    model_mapping: null,
  }
  const [inputs, setInputs] = useState(originInputs);


  const handleSave = async () => {
    setLoading(true);
    let data = {
      tag: tag,
    }
    if (inputs.newTag === tag) {
      setLoading(false);
      return;
    }
    if (inputs.model_mapping !== null) {
      data.model_mapping = inputs.model
    }
    data.newTag = inputs.newTag;
    if (data.newTag === '') {
      Modal.confirm({
        title: '解散标签',
        content: '确定要解散标签吗？',
        onCancel: () => {
          setLoading(false);
        },
        onOk: async () => {
          await submit(data);
        }
      });
    } else {
      await submit(data);
    }
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
  }

  useEffect(() => {
    setInputs({
      ...originInputs,
      tag: tag,
      newTag: tag,
    })
  }, [visible]);

  return (
    <SideSheet
      title="编辑标签"
      visible={visible}
      onCancel={handleClose}
      footer={
        <div style={{ display: 'flex', justifyContent: 'flex-end' }}>
          <Space>
            <Button onClick={handleClose}>取消</Button>
            <Button type="primary" onClick={handleSave} loading={loading}>保存</Button>
          </Space>
        </div>
      }
    >
      <Spin spinning={loading}>
        <TextInput
          label="新标签（留空则解散标签，不会删除标签下的渠道）"
          name="newTag"
          value={inputs.new_tag}
          onChange={(value) => setInputs({ ...inputs, new_tag: value })}
          placeholder="请输入新标签"
        />
      </Spin>
    </SideSheet>
  );
};

export default EditTagModal;