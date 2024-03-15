import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { API, isMobile, showError, showSuccess } from '../../helpers';
import { renderQuotaWithPrompt } from '../../helpers/render';
import Title from '@douyinfe/semi-ui/lib/es/typography/title';
import { Button, Divider, Input, Select, SideSheet, Space, Spin, Typography } from '@douyinfe/semi-ui';

const EditUser = (props) => {
  const userId = props.editingUser.id;
  const [loading, setLoading] = useState(true);
  const [inputs, setInputs] = useState({
    username: '',
    display_name: '',
    password: '',
    github_id: '',
    wechat_id: '',
    email: '',
    quota: 0,
    group: 'default'
  });
  const [groupOptions, setGroupOptions] = useState([]);
  const { username, display_name, password, github_id, wechat_id, telegram_id, email, quota, group } =
    inputs;
  const handleInputChange = (name, value) => {
    setInputs((inputs) => ({ ...inputs, [name]: value }));
  };
  const fetchGroups = async () => {
    try {
      let res = await API.get(`/api/group/`);
      setGroupOptions(res.data.data.map((group) => ({
        label: group,
        value: group
      })));
    } catch (error) {
      showError(error.message);
    }
  };
  const navigate = useNavigate();
  const handleCancel = () => {
    props.handleClose();
  };
  const loadUser = async () => {
    setLoading(true);
    let res = undefined;
    if (userId) {
      res = await API.get(`/api/user/${userId}`);
    } else {
      res = await API.get(`/api/user/self`);
    }
    const { success, message, data } = res.data;
    if (success) {
      data.password = '';
      setInputs(data);
    } else {
      showError(message);
    }
    setLoading(false);
  };

  useEffect(() => {
    loadUser().then();
    if (userId) {
      fetchGroups().then();
    }
  }, [props.editingUser.id]);

  const submit = async () => {
    setLoading(true);
    let res = undefined;
    if (userId) {
      let data = { ...inputs, id: parseInt(userId) };
      if (typeof data.quota === 'string') {
        data.quota = parseInt(data.quota);
      }
      res = await API.put(`/api/user/`, data);
    } else {
      res = await API.put(`/api/user/self`, inputs);
    }
    const { success, message } = res.data;
    if (success) {
      showSuccess('用户信息更新成功！');
      props.refresh();
      props.handleClose();
    } else {
      showError(message);
    }
    setLoading(false);
  };

  return (
    <>
      <SideSheet
        placement={'right'}
        title={<Title level={3}>{'编辑用户'}</Title>}
        headerStyle={{ borderBottom: '1px solid var(--semi-color-border)' }}
        bodyStyle={{ borderBottom: '1px solid var(--semi-color-border)' }}
        visible={props.visible}
        footer={
          <div style={{ display: 'flex', justifyContent: 'flex-end' }}>
            <Space>
              <Button theme="solid" size={'large'} onClick={submit}>提交</Button>
              <Button theme="solid" size={'large'} type={'tertiary'} onClick={handleCancel}>取消</Button>
            </Space>
          </div>
        }
        closeIcon={null}
        onCancel={() => handleCancel()}
        width={isMobile() ? '100%' : 600}
      >
        <Spin spinning={loading}>
          <div style={{ marginTop: 20 }}>
            <Typography.Text>用户名</Typography.Text>
          </div>
          <Input
            label="用户名"
            name="username"
            placeholder={'请输入新的用户名'}
            onChange={value => handleInputChange('username', value)}
            value={username}
            autoComplete="new-password"
          />
          <div style={{ marginTop: 20 }}>
            <Typography.Text>密码</Typography.Text>
          </div>
          <Input
            label="密码"
            name="password"
            type={'password'}
            placeholder={'请输入新的密码，最短 8 位'}
            onChange={value => handleInputChange('password', value)}
            value={password}
            autoComplete="new-password"
          />
          <div style={{ marginTop: 20 }}>
            <Typography.Text>显示名称</Typography.Text>
          </div>
          <Input
            label="显示名称"
            name="display_name"
            placeholder={'请输入新的显示名称'}
            onChange={value => handleInputChange('display_name', value)}
            value={display_name}
            autoComplete="new-password"
          />
          {
            userId && <>
              <div style={{ marginTop: 20 }}>
                <Typography.Text>分组</Typography.Text>
              </div>
              <Select
                placeholder={'请选择分组'}
                name="group"
                fluid
                search
                selection
                allowAdditions
                additionLabel={'请在系统设置页面编辑分组倍率以添加新的分组：'}
                onChange={value => handleInputChange('group', value)}
                value={inputs.group}
                autoComplete="new-password"
                optionList={groupOptions}
              />
              <div style={{ marginTop: 20 }}>
                <Typography.Text>{`剩余额度${renderQuotaWithPrompt(quota)}`}</Typography.Text>
              </div>
              <Input
                name="quota"
                placeholder={'请输入新的剩余额度'}
                onChange={value => handleInputChange('quota', value)}
                value={quota}
                type={'number'}
                autoComplete="new-password"
              />
            </>
          }
          <Divider style={{ marginTop: 20 }}>以下信息不可修改</Divider>
          <div style={{ marginTop: 20 }}>
            <Typography.Text>已绑定的 GitHub 账户</Typography.Text>
          </div>
          <Input
            name="github_id"
            value={github_id}
            autoComplete="new-password"
            placeholder="此项只读，需要用户通过个人设置页面的相关绑定按钮进行绑定，不可直接修改"
            readonly
          />
          <div style={{ marginTop: 20 }}>
            <Typography.Text>已绑定的微信账户</Typography.Text>
          </div>
          <Input
            name="wechat_id"
            value={wechat_id}
            autoComplete="new-password"
            placeholder="此项只读，需要用户通过个人设置页面的相关绑定按钮进行绑定，不可直接修改"
            readonly
          />
          <Input
            name="telegram_id"
            value={telegram_id}
            autoComplete="new-password"
            placeholder="此项只读，需要用户通过个人设置页面的相关绑定按钮进行绑定，不可直接修改"
            readonly
          />
          <div style={{ marginTop: 20 }}>
            <Typography.Text>已绑定的邮箱账户</Typography.Text>
          </div>
          <Input
            name="email"
            value={email}
            autoComplete="new-password"
            placeholder="此项只读，需要用户通过个人设置页面的相关绑定按钮进行绑定，不可直接修改"
            readonly
          />
        </Spin>
      </SideSheet>
    </>
  );
};

export default EditUser;
