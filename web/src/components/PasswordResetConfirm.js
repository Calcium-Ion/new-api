import React, { useEffect, useState } from 'react';
import { Button, Form, Grid, Header, Image, Segment } from 'semantic-ui-react';
import { API, copy, showError, showNotice } from '../helpers';
import { useSearchParams } from 'react-router-dom';

const PasswordResetConfirm = () => {
  const [inputs, setInputs] = useState({
    email: '',
    token: ''
  });
  const { email, token } = inputs;

  const [loading, setLoading] = useState(false);

  const [disableButton, setDisableButton] = useState(false);
  const [countdown, setCountdown] = useState(30);

  const [newPassword, setNewPassword] = useState('');

  const [searchParams, setSearchParams] = useSearchParams();
  useEffect(() => {
    let token = searchParams.get('token');
    let email = searchParams.get('email');
    setInputs({
      token,
      email
    });
  }, []);

  useEffect(() => {
    let countdownInterval = null;
    if (disableButton && countdown > 0) {
      countdownInterval = setInterval(() => {
        setCountdown(countdown - 1);
      }, 1000);
    } else if (countdown === 0) {
      setDisableButton(false);
      setCountdown(30);
    }
    return () => clearInterval(countdownInterval);
  }, [disableButton, countdown]);

  async function handleSubmit(e) {
    setDisableButton(true);
    if (!email) return;
    setLoading(true);
    const res = await API.post(`/api/user/reset`, {
      email,
      token
    });
    const { success, message } = res.data;
    if (success) {
      let password = res.data.data;
      setNewPassword(password);
      await copy(password);
      showNotice(`新密码已复制到剪贴板：${password}`);
    } else {
      showError(message);
    }
    setLoading(false);
  }

  return (
    <Grid textAlign="center" style={{ marginTop: '48px' }}>
      <Grid.Column style={{ maxWidth: 450 }}>
        <Header as="h2" color="" textAlign="center">
          <Image src="/logo.png" /> 密码重置确认
        </Header>
        <Form size="large">
          <Segment>
            <Form.Input
              fluid
              icon="mail"
              iconPosition="left"
              placeholder="邮箱地址"
              name="email"
              value={email}
              readOnly
            />
            {newPassword && (
              <Form.Input
                fluid
                icon="lock"
                iconPosition="left"
                placeholder="新密码"
                name="newPassword"
                value={newPassword}
                readOnly
                onClick={(e) => {
                  e.target.select();
                  navigator.clipboard.writeText(newPassword);
                  showNotice(`密码已复制到剪贴板：${newPassword}`);
                }}
              />
            )}
            <Button
              color="green"
              fluid
              size="large"
              onClick={handleSubmit}
              loading={loading}
              disabled={disableButton}
            >
              {disableButton ? `密码重置完成` : '提交'}
            </Button>
          </Segment>
        </Form>
      </Grid.Column>
    </Grid>
  );
};

export default PasswordResetConfirm;
