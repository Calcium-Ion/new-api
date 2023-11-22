import React, { useEffect, useState } from 'react';
import { Button, Form, Header, Segment } from 'semantic-ui-react';
import { useParams, useNavigate } from 'react-router-dom';
import { API, showError, showSuccess } from '../../helpers';
import { renderQuota, renderQuotaWithPrompt } from '../../helpers/render';

const EditUser = () => {
  const params = useParams();
  const userId = params.id;
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
  const { username, display_name, password, github_id, wechat_id, email, quota, group } =
      inputs;
  const handleInputChange = (e, { name, value }) => {
    setInputs((inputs) => ({ ...inputs, [name]: value }));
  };
  const fetchGroups = async () => {
    try {
      let res = await API.get(`/api/group/`);
      setGroupOptions(res.data.data.map((group) => ({
        key: group,
        text: group,
        value: group,
      })));
    } catch (error) {
      showError(error.message);
    }
  };
  const navigate = useNavigate();
  const handleCancel = () => {
    navigate("/setting");
  }
  const loadUser = async () => {
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
  }, []);

  const submit = async () => {
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
    } else {
      showError(message);
    }
  };

  return (
      <>
        <Segment loading={loading}>
          <Header as='h3'>更新用户信息</Header>
          <Form autoComplete='new-password'>
            <Form.Field>
              <Form.Input
                  label='用户名'
                  name='username'
                  placeholder={'请输入新的用户名'}
                  onChange={handleInputChange}
                  value={username}
                  autoComplete='new-password'
              />
            </Form.Field>
            <Form.Field>
              <Form.Input
                  label='密码'
                  name='password'
                  type={'password'}
                  placeholder={'请输入新的密码，最短 8 位'}
                  onChange={handleInputChange}
                  value={password}
                  autoComplete='new-password'
              />
            </Form.Field>
            <Form.Field>
              <Form.Input
                  label='显示名称'
                  name='display_name'
                  placeholder={'请输入新的显示名称'}
                  onChange={handleInputChange}
                  value={display_name}
                  autoComplete='new-password'
              />
            </Form.Field>
            {
                userId && <>
                  <Form.Field>
                    <Form.Dropdown
                        label='分组'
                        placeholder={'请选择分组'}
                        name='group'
                        fluid
                        search
                        selection
                        allowAdditions
                        additionLabel={'请在系统设置页面编辑分组倍率以添加新的分组：'}
                        onChange={handleInputChange}
                        value={inputs.group}
                        autoComplete='new-password'
                        options={groupOptions}
                    />
                  </Form.Field>
                  <Form.Field>
                    <Form.Input
                        label={`剩余额度${renderQuotaWithPrompt(quota)}`}
                        name='quota'
                        placeholder={'请输入新的剩余额度'}
                        onChange={handleInputChange}
                        value={quota}
                        type={'number'}
                        autoComplete='new-password'
                    />
                  </Form.Field>
                </>
            }
            <Form.Field>
              <Form.Input
                  label='已绑定的 GitHub 账户'
                  name='github_id'
                  value={github_id}
                  autoComplete='new-password'
                  placeholder='此项只读，需要用户通过个人设置页面的相关绑定按钮进行绑定，不可直接修改'
                  readOnly
              />
            </Form.Field>
            <Form.Field>
              <Form.Input
                  label='已绑定的微信账户'
                  name='wechat_id'
                  value={wechat_id}
                  autoComplete='new-password'
                  placeholder='此项只读，需要用户通过个人设置页面的相关绑定按钮进行绑定，不可直接修改'
                  readOnly
              />
            </Form.Field>
            <Form.Field>
              <Form.Input
                  label='已绑定的邮箱账户'
                  name='email'
                  value={email}
                  autoComplete='new-password'
                  placeholder='此项只读，需要用户通过个人设置页面的相关绑定按钮进行绑定，不可直接修改'
                  readOnly
              />
            </Form.Field>
            <Button onClick={handleCancel}>取消</Button>
            <Button positive onClick={submit}>提交</Button>
          </Form>
        </Segment>
      </>
  );
};

export default EditUser;
