import React, { useContext, useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import {
  API,
  copy,
  isRoot,
  showError,
  showInfo,
  showSuccess
} from '../helpers';
import Turnstile from 'react-turnstile';
import { UserContext } from '../context/User';
import { onGitHubOAuthClicked, onOIDCClicked, onLinuxDOOAuthClicked } from './utils';
import {
  Avatar,
  Banner,
  Button,
  Card,
  Descriptions,
  Image,
  Input,
  InputNumber,
  Layout,
  Modal,
  Space,
  Tag,
  Typography,
  Collapsible,
  Select,
  Radio,
  RadioGroup,
  AutoComplete,
  Checkbox,
  Tabs,
  TabPane
} from '@douyinfe/semi-ui';
import {
  getQuotaPerUnit,
  renderQuota,
  renderQuotaWithPrompt,
  stringToColor
} from '../helpers/render';
import TelegramLoginButton from 'react-telegram-login';
import { useTranslation } from 'react-i18next';

const PersonalSetting = () => {
  const [userState, userDispatch] = useContext(UserContext);
  let navigate = useNavigate();
  const { t } = useTranslation();

  const [inputs, setInputs] = useState({
    wechat_verification_code: '',
    email_verification_code: '',
    email: '',
    self_account_deletion_confirmation: '',
    set_new_password: '',
    set_new_password_confirmation: ''
  });
  const [status, setStatus] = useState({});
  const [showChangePasswordModal, setShowChangePasswordModal] = useState(false);
  const [showWeChatBindModal, setShowWeChatBindModal] = useState(false);
  const [showEmailBindModal, setShowEmailBindModal] = useState(false);
  const [showAccountDeleteModal, setShowAccountDeleteModal] = useState(false);
  const [turnstileEnabled, setTurnstileEnabled] = useState(false);
  const [turnstileSiteKey, setTurnstileSiteKey] = useState('');
  const [turnstileToken, setTurnstileToken] = useState('');
  const [loading, setLoading] = useState(false);
  const [disableButton, setDisableButton] = useState(false);
  const [countdown, setCountdown] = useState(30);
  const [affLink, setAffLink] = useState('');
  const [systemToken, setSystemToken] = useState('');
  const [models, setModels] = useState([]);
  const [openTransfer, setOpenTransfer] = useState(false);
  const [transferAmount, setTransferAmount] = useState(0);
  const [isModelsExpanded, setIsModelsExpanded] = useState(() => {
    // Initialize from localStorage if available
    const savedState = localStorage.getItem('modelsExpanded');
    return savedState ? JSON.parse(savedState) : false;
  });
  const MODELS_DISPLAY_COUNT = 10;  // 默认显示的模型数量
  const [notificationSettings, setNotificationSettings] = useState({
    warningType: 'email',
    warningThreshold: 100000,
    webhookUrl: '',
    webhookSecret: '',
    notificationEmail: '',
    acceptUnsetModelRatioModel: false
  });
  const [showWebhookDocs, setShowWebhookDocs] = useState(false);

  useEffect(() => {
    let status = localStorage.getItem('status');
    if (status) {
      status = JSON.parse(status);
      setStatus(status);
      if (status.turnstile_check) {
        setTurnstileEnabled(true);
        setTurnstileSiteKey(status.turnstile_site_key);
      }
    }
    getUserData().then((res) => {
      console.log(userState);
    });
    loadModels().then();
    getAffLink().then();
    setTransferAmount(getQuotaPerUnit());
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
    return () => clearInterval(countdownInterval); // Clean up on unmount
  }, [disableButton, countdown]);

  useEffect(() => {
    if (userState?.user?.setting) {
      const settings = JSON.parse(userState.user.setting);
      setNotificationSettings({
        warningType: settings.notify_type || 'email',
        warningThreshold: settings.quota_warning_threshold || 500000,
        webhookUrl: settings.webhook_url || '',
        webhookSecret: settings.webhook_secret || '',
        notificationEmail: settings.notification_email || '',
        acceptUnsetModelRatioModel: settings.accept_unset_model_ratio_model || false
      });
    }
  }, [userState?.user?.setting]);

  // Save models expanded state to localStorage whenever it changes
  useEffect(() => {
    localStorage.setItem('modelsExpanded', JSON.stringify(isModelsExpanded));
  }, [isModelsExpanded]);

  const handleInputChange = (name, value) => {
    setInputs((inputs) => ({ ...inputs, [name]: value }));
  };

  const generateAccessToken = async () => {
    const res = await API.get('/api/user/token');
    const { success, message, data } = res.data;
    if (success) {
      setSystemToken(data);
      await copy(data);
      showSuccess(t('令牌已重置并已复制到剪贴板'));
    } else {
      showError(message);
    }
  };

  const getAffLink = async () => {
    const res = await API.get('/api/user/aff');
    const { success, message, data } = res.data;
    if (success) {
      let link = `${window.location.origin}/register?aff=${data}`;
      setAffLink(link);
    } else {
      showError(message);
    }
  };

  const getUserData = async () => {
    let res = await API.get(`/api/user/self`);
    const { success, message, data } = res.data;
    if (success) {
      userDispatch({ type: 'login', payload: data });
    } else {
      showError(message);
    }
  };

  const loadModels = async () => {
    let res = await API.get(`/api/user/models`);
    const { success, message, data } = res.data;
    if (success) {
      if (data != null) {
        setModels(data);
      }
    } else {
      showError(message);
    }
  };

  const handleAffLinkClick = async (e) => {
    e.target.select();
    await copy(e.target.value);
    showSuccess(t('邀请链接已复制到剪切板'));
  };

  const handleSystemTokenClick = async (e) => {
    e.target.select();
    await copy(e.target.value);
    showSuccess(t('系统令牌已复制到剪切板'));
  };

  const deleteAccount = async () => {
    if (inputs.self_account_deletion_confirmation !== userState.user.username) {
      showError(t('请输入你的账户名以确认删除！'));
      return;
    }

    const res = await API.delete('/api/user/self');
    const { success, message } = res.data;

    if (success) {
      showSuccess(t('账户已删除！'));
      await API.get('/api/user/logout');
      userDispatch({ type: 'logout' });
      localStorage.removeItem('user');
      navigate('/login');
    } else {
      showError(message);
    }
  };

  const bindWeChat = async () => {
    if (inputs.wechat_verification_code === '') return;
    const res = await API.get(
      `/api/oauth/wechat/bind?code=${inputs.wechat_verification_code}`
    );
    const { success, message } = res.data;
    if (success) {
      showSuccess(t('微信账户绑定成功！'));
      setShowWeChatBindModal(false);
    } else {
      showError(message);
    }
  };

  const changePassword = async () => {
    if (inputs.set_new_password !== inputs.set_new_password_confirmation) {
      showError(t('两次输入的密码不一致！'));
      return;
    }
    const res = await API.put(`/api/user/self`, {
      password: inputs.set_new_password
    });
    const { success, message } = res.data;
    if (success) {
      showSuccess(t('密码修改成功！'));
      setShowWeChatBindModal(false);
    } else {
      showError(message);
    }
    setShowChangePasswordModal(false);
  };

  const transfer = async () => {
    if (transferAmount < getQuotaPerUnit()) {
      showError(t('划转金额最低为') + ' ' + renderQuota(getQuotaPerUnit()));
      return;
    }
    const res = await API.post(`/api/user/aff_transfer`, {
      quota: transferAmount
    });
    const { success, message } = res.data;
    if (success) {
      showSuccess(message);
      setOpenTransfer(false);
      getUserData().then();
    } else {
      showError(message);
    }
  };

  const sendVerificationCode = async () => {
    if (inputs.email === '') {
      showError(t('请输入邮箱！'));
      return;
    }
    setDisableButton(true);
    if (turnstileEnabled && turnstileToken === '') {
      showInfo('请稍后几秒重试，Turnstile 正在检查用户环境！');
      return;
    }
    setLoading(true);
    const res = await API.get(
      `/api/verification?email=${inputs.email}&turnstile=${turnstileToken}`
    );
    const { success, message } = res.data;
    if (success) {
      showSuccess(t('验证码发送成功，请检查邮箱！'));
    } else {
      showError(message);
    }
    setLoading(false);
  };

  const bindEmail = async () => {
    if (inputs.email_verification_code === '') {
      showError(t('请输入邮箱验证码！'));
      return;
    }
    setLoading(true);
    const res = await API.get(
      `/api/oauth/email/bind?email=${inputs.email}&code=${inputs.email_verification_code}`
    );
    const { success, message } = res.data;
    if (success) {
      showSuccess(t('邮箱账户绑定成功！'));
      setShowEmailBindModal(false);
      userState.user.email = inputs.email;
    } else {
      showError(message);
    }
    setLoading(false);
  };

  const getUsername = () => {
    if (userState.user) {
      return userState.user.username;
    } else {
      return 'null';
    }
  };

  const handleCancel = () => {
    setOpenTransfer(false);
  };

  const copyText = async (text) => {
    if (await copy(text)) {
      showSuccess(t('已复制：') + text);
    } else {
      // setSearchKeyword(text);
      Modal.error({ title: t('无法复制到剪贴板，请手动复制'), content: text });
    }
  };

  const handleNotificationSettingChange = (type, value) => {
    setNotificationSettings(prev => ({
      ...prev,
      [type]: value.target ? value.target.value : value  // 处理 Radio 事件对象
    }));
  };

  const saveNotificationSettings = async () => {
    try {
      const res = await API.put('/api/user/setting', {
        notify_type: notificationSettings.warningType,
        quota_warning_threshold: parseFloat(notificationSettings.warningThreshold),
        webhook_url: notificationSettings.webhookUrl,
        webhook_secret: notificationSettings.webhookSecret,
        notification_email: notificationSettings.notificationEmail,
        accept_unset_model_ratio_model: notificationSettings.acceptUnsetModelRatioModel
      });

      if (res.data.success) {
        showSuccess(t('通知设置已更新'));
        await getUserData();
      } else {
        showError(res.data.message);
      }
    } catch (error) {
      showError(t('更新通知设置失败'));
    }
  };

  return (

    <div>
      <Layout>
        <Layout.Content>
          <Modal
            title={t('请输入要划转的数量')}
            visible={openTransfer}
            onOk={transfer}
            onCancel={handleCancel}
            maskClosable={false}
            size={'small'}
            centered={true}
          >
            <div style={{ marginTop: 20 }}>
              <Typography.Text>{t('可用额度')}{renderQuotaWithPrompt(userState?.user?.aff_quota)}</Typography.Text>
              <Input
                style={{ marginTop: 5 }}
                value={userState?.user?.aff_quota}
                disabled={true}
              ></Input>
            </div>
            <div style={{ marginTop: 20 }}>
              <Typography.Text>
                {t('划转额度')}{renderQuotaWithPrompt(transferAmount)} {t('最低') + renderQuota(getQuotaPerUnit())}
              </Typography.Text>
              <div>
                <InputNumber
                  min={0}
                  style={{ marginTop: 5 }}
                  value={transferAmount}
                  onChange={(value) => setTransferAmount(value)}
                  disabled={false}
                ></InputNumber>
              </div>
            </div>
          </Modal>
          <div>
            <Card
              title={
                <Card.Meta
                  avatar={
                    <Avatar
                      size="default"
                      color={stringToColor(getUsername())}
                      style={{ marginRight: 4 }}
                    >
                      {typeof getUsername() === 'string' &&
                        getUsername().slice(0, 1)}
                    </Avatar>
                  }
                  title={<Typography.Text>{getUsername()}</Typography.Text>}
                  description={
                    isRoot() ? (
                      <Tag color="red">{t('管理员')}</Tag>
                    ) : (
                      <Tag color="blue">{t('普通用户')}</Tag>
                    )
                  }
                ></Card.Meta>
              }
              headerExtraContent={
                <>
                  <Space vertical align="start">
                    <Tag color="green">{'ID: ' + userState?.user?.id}</Tag>
                    <Tag color="blue">{userState?.user?.group}</Tag>
                  </Space>
                </>
              }
              footer={
                <>
                  <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                    <Typography.Title heading={6}>{t('可用模型')}</Typography.Title>
                  </div>
                  <div style={{ marginTop: 10 }}>
                    {models.length <= MODELS_DISPLAY_COUNT ? (
                      <Space wrap>
                        {models.map((model) => (
                          <Tag
                            key={model}
                            color="cyan"
                            onClick={() => {
                              copyText(model);
                            }}
                          >
                            {model}
                          </Tag>
                        ))}
                      </Space>
                    ) : (
                      <>
                        <Collapsible isOpen={isModelsExpanded}>
                          <Space wrap>
                            {models.map((model) => (
                              <Tag
                                key={model}
                                color="cyan"
                                onClick={() => {
                                  copyText(model);
                                }}
                              >
                                {model}
                              </Tag>
                            ))}
                            <Tag
                              color="blue"
                              type="light"
                              style={{ cursor: 'pointer' }}
                              onClick={() => setIsModelsExpanded(false)}
                            >
                              {t('收起')}
                            </Tag>
                          </Space>
                        </Collapsible>
                        {!isModelsExpanded && (
                          <Space wrap>
                            {models.slice(0, MODELS_DISPLAY_COUNT).map((model) => (
                              <Tag
                                key={model}
                                color="cyan"
                                onClick={() => {
                                  copyText(model);
                                }}
                              >
                                {model}
                              </Tag>
                            ))}
                            <Tag
                              color="blue"
                              type="light"
                              style={{ cursor: 'pointer' }}
                              onClick={() => setIsModelsExpanded(true)}
                            >
                              {t('更多')} {models.length - MODELS_DISPLAY_COUNT} {t('个模型')}
                            </Tag>
                          </Space>
                        )}
                      </>
                    )}
                  </div>
                </>

              }
            >
              <Descriptions row>
                <Descriptions.Item itemKey={t('当前余额')}>
                  {renderQuota(userState?.user?.quota)}
                </Descriptions.Item>
                <Descriptions.Item itemKey={t('历史消耗')}>
                  {renderQuota(userState?.user?.used_quota)}
                </Descriptions.Item>
                <Descriptions.Item itemKey={t('请求次数')}>
                  {userState.user?.request_count}
                </Descriptions.Item>
              </Descriptions>
            </Card>
            <Card
              style={{ marginTop: 10 }}
              footer={
                <div>
                  <Typography.Text>{t('邀请链接')}</Typography.Text>
                  <Input
                    style={{ marginTop: 10 }}
                    value={affLink}
                    onClick={handleAffLinkClick}
                    readOnly
                  />
                </div>
              }
            >
              <Typography.Title heading={6}>{t('邀请信息')}</Typography.Title>
              <div style={{ marginTop: 10 }}>
                <Descriptions row>
                  <Descriptions.Item itemKey={t('待使用收益')}>
                                        <span style={{ color: 'rgba(var(--semi-red-5), 1)' }}>
                                            {renderQuota(userState?.user?.aff_quota)}
                                        </span>
                    <Button
                      type={'secondary'}
                      onClick={() => setOpenTransfer(true)}
                      size={'small'}
                      style={{ marginLeft: 10 }}
                    >
                      {t('划转')}
                    </Button>
                  </Descriptions.Item>
                  <Descriptions.Item itemKey={t('总收益')}>
                    {renderQuota(userState?.user?.aff_history_quota)}
                  </Descriptions.Item>
                  <Descriptions.Item itemKey={t('邀请人数')}>
                    {userState?.user?.aff_count}
                  </Descriptions.Item>
                </Descriptions>
              </div>
            </Card>
            <Card style={{ marginTop: 10 }}>
              <Typography.Title heading={6}>{t('个人信息')}</Typography.Title>
              <div style={{ marginTop: 20 }}>
                <Typography.Text strong>{t('邮箱')}</Typography.Text>
                <div
                  style={{ display: 'flex', justifyContent: 'space-between' }}
                >
                  <div>
                    <Input
                      value={
                        userState.user && userState.user.email !== ''
                          ? userState.user.email
                          : t('未绑定')
                      }
                      readonly={true}
                    ></Input>
                  </div>
                  <div>
                    <Button
                      onClick={() => {
                        setShowEmailBindModal(true);
                      }}
                    >
                      {userState.user && userState.user.email !== ''
                        ? t('修改绑定')
                        : t('绑定邮箱')}
                    </Button>
                  </div>
                </div>
              </div>
              <div style={{ marginTop: 10 }}>
                <Typography.Text strong>{t('微信')}</Typography.Text>
                <div style={{ display: 'flex', justifyContent: 'space-between' }}>
                  <div>
                    <Input
                      value={
                        userState.user && userState.user.wechat_id !== ''
                          ? t('已绑定')
                          : t('未绑定')
                      }
                      readonly={true}
                    ></Input>
                  </div>
                  <div>
                    <Button
                      disabled={!status.wechat_login}
                      onClick={() => {
                        setShowWeChatBindModal(true);
                      }}
                    >
                      {userState.user && userState.user.wechat_id !== ''
                        ? t('修改绑定')
                        : status.wechat_login
                          ? t('绑定')
                          : t('未启用')}
                    </Button>
                  </div>
                </div>
              </div>
              <div style={{ marginTop: 10 }}>
                <Typography.Text strong>{t('GitHub')}</Typography.Text>
                <div
                  style={{ display: 'flex', justifyContent: 'space-between' }}
                >
                  <div>
                    <Input
                      value={
                        userState.user && userState.user.github_id !== ''
                          ? userState.user.github_id
                          : t('未绑定')
                      }
                      readonly={true}
                    ></Input>
                  </div>
                  <div>
                    <Button
                      onClick={() => {
                        onGitHubOAuthClicked(status.github_client_id);
                      }}
                      disabled={
                        (userState.user && userState.user.github_id !== '') ||
                        !status.github_oauth
                      }
                    >
                      {status.github_oauth ? t('绑定') : t('未启用')}
                    </Button>
                  </div>
                </div>
              </div>
              <div style={{ marginTop: 10 }}>
                <Typography.Text strong>{t('OIDC')}</Typography.Text>
                <div
                  style={{ display: 'flex', justifyContent: 'space-between' }}
                >
                  <div>
                    <Input
                      value={
                        userState.user && userState.user.oidc_id !== ''
                          ? userState.user.oidc_id
                          : t('未绑定')
                      }
                      readonly={true}
                    ></Input>
                  </div>
                  <div>
                    <Button
                      onClick={() => {
                        onOIDCClicked(status.oidc_authorization_endpoint, status.oidc_client_id);
                      }}
                      disabled={
                        (userState.user && userState.user.oidc_id !== '') ||
                        !status.oidc_enabled
                      }
                    >
                      {status.oidc_enabled ? t('绑定') : t('未启用')}
                    </Button>
                  </div>
                </div>
              </div>
              <div style={{ marginTop: 10 }}>
                <Typography.Text strong>{t('Telegram')}</Typography.Text>
                <div
                  style={{ display: 'flex', justifyContent: 'space-between' }}
                >
                  <div>
                    <Input
                      value={
                        userState.user && userState.user.telegram_id !== ''
                          ? userState.user.telegram_id
                          : t('未绑定')
                      }
                      readonly={true}
                    ></Input>
                  </div>
                  <div>
                    {status.telegram_oauth ? (
                      userState.user.telegram_id !== '' ? (
                        <Button disabled={true}>{t('已绑定')}</Button>
                      ) : (
                        <TelegramLoginButton
                          dataAuthUrl="/api/oauth/telegram/bind"
                          botName={status.telegram_bot_name}
                        />
                      )
                    ) : (
                      <Button disabled={true}>{t('未启用')}</Button>
                    )}
                  </div>
                </div>
              </div>
              <div style={{ marginTop: 10 }}>
                <Typography.Text strong>{t('LinuxDO')}</Typography.Text>
                <div
                  style={{ display: 'flex', justifyContent: 'space-between' }}
                >
                  <div>
                    <Input
                      value={
                        userState.user && userState.user.linux_do_id !== ''
                          ? userState.user.linux_do_id
                          : t('未绑定')
                      }
                      readonly={true}
                    ></Input>
                  </div>
                  <div>
                    <Button
                      onClick={() => {
                        onLinuxDOOAuthClicked(status.linuxdo_client_id);
                      }}
                      disabled={
                        (userState.user && userState.user.linux_do_id !== '') ||
                        !status.linuxdo_oauth
                      }
                    >
                      {status.linuxdo_oauth ? t('绑定') : t('未启用')}
                    </Button>
                  </div>
                </div>
              </div>
              <div style={{ marginTop: 10 }}>
                <Space>
                  <Button onClick={generateAccessToken}>
                    {t('生成系统访问令牌')}
                  </Button>
                  <Button
                    onClick={() => {
                      setShowChangePasswordModal(true);
                    }}
                  >
                    {t('修改密码')}
                  </Button>
                  <Button
                    type={'danger'}
                    onClick={() => {
                      setShowAccountDeleteModal(true);
                    }}
                  >
                    {t('删除个人账户')}
                  </Button>
                </Space>

                {systemToken && (
                  <Input
                    readOnly
                    value={systemToken}
                    onClick={handleSystemTokenClick}
                    style={{ marginTop: '10px' }}
                  />
                )}
                <Modal
                  onCancel={() => setShowWeChatBindModal(false)}
                  visible={showWeChatBindModal}
                  size={'small'}
                >
                  <Image src={status.wechat_qrcode} />
                  <div style={{ textAlign: 'center' }}>
                    <p>
                      微信扫码关注公众号，输入「验证码」获取验证码（三分钟内有效）
                    </p>
                  </div>
                  <Input
                    placeholder="验证码"
                    name="wechat_verification_code"
                    value={inputs.wechat_verification_code}
                    onChange={(v) =>
                      handleInputChange('wechat_verification_code', v)
                    }
                  />
                  <Button color="" fluid size="large" onClick={bindWeChat}>
                    {t('绑定')}
                  </Button>
                </Modal>
              </div>
            </Card>
            <Card style={{ marginTop: 10 }}>
              <Tabs type="line" defaultActiveKey="price">
                <TabPane tab={t('价格设置')} itemKey="price">
                  <div style={{ marginTop: 20 }}>
                    <Typography.Text strong>{t('接受未设置价格模型')}</Typography.Text>
                    <div style={{ marginTop: 10 }}>
                      <Checkbox
                        checked={notificationSettings.acceptUnsetModelRatioModel}
                        onChange={e => handleNotificationSettingChange('acceptUnsetModelRatioModel', e.target.checked)}
                      >
                        {t('接受未设置价格模型')}
                      </Checkbox>
                      <Typography.Text type="secondary" style={{ marginTop: 8, display: 'block' }}>
                        {t('当模型没有设置价格时仍接受调用，仅当您信任该网站时使用，可能会产生高额费用')}
                      </Typography.Text>
                    </div>
                  </div>
                </TabPane>
                <TabPane tab={t('通知设置')} itemKey="notification">
                  <div style={{ marginTop: 20 }}>
                    <Typography.Text strong>{t('通知方式')}</Typography.Text>
                    <div style={{ marginTop: 10 }}>
                      <RadioGroup
                        value={notificationSettings.warningType}
                        onChange={value => handleNotificationSettingChange('warningType', value)}
                      >
                        <Radio value="email">{t('邮件通知')}</Radio>
                        <Radio value="webhook">{t('Webhook通知')}</Radio>
                      </RadioGroup>
                    </div>
                  </div>
                  {notificationSettings.warningType === 'webhook' && (
                    <>
                      <div style={{ marginTop: 20 }}>
                        <Typography.Text strong>{t('Webhook地址')}</Typography.Text>
                        <div style={{ marginTop: 10 }}>
                          <Input
                            value={notificationSettings.webhookUrl}
                            onChange={val => handleNotificationSettingChange('webhookUrl', val)}
                            placeholder={t('请输入Webhook地址，例如: https://example.com/webhook')}
                          />
                          <Typography.Text type="secondary" style={{ marginTop: 8, display: 'block' }}>
                            {t('只支持https，系统将以 POST 方式发送通知，请确保地址可以接收 POST 请求')}
                          </Typography.Text>
                          <Typography.Text type="secondary" style={{ marginTop: 8, display: 'block' }}>
                            <div style={{ cursor: 'pointer' }} onClick={() => setShowWebhookDocs(!showWebhookDocs)}>
                              {t('Webhook请求结构')} {showWebhookDocs ? '▼' : '▶'}
                            </div>
                            <Collapsible isOpen={showWebhookDocs}>
                            <pre style={{
                              marginTop: 4,
                              background: 'var(--semi-color-fill-0)',
                              padding: 8,
                              borderRadius: 4
                            }}>
{`{
    "type": "quota_exceed",      // 通知类型
    "title": "标题",             // 通知标题
    "content": "通知内容",       // 通知内容，支持 {{value}} 变量占位符
    "values": ["值1", "值2"],    // 按顺序替换content中的 {{value}} 占位符
    "timestamp": 1739950503      // 时间戳
}

示例：
{
    "type": "quota_exceed",
    "title": "额度预警通知",
    "content": "您的额度即将用尽，当前剩余额度为 {{value}}",
    "values": ["$0.99"],
    "timestamp": 1739950503
}`}
                            </pre>
                            </Collapsible>
                          </Typography.Text>
                        </div>
                      </div>
                      <div style={{ marginTop: 20 }}>
                        <Typography.Text strong>{t('接口凭证（可选）')}</Typography.Text>
                        <div style={{ marginTop: 10 }}>
                          <Input
                            value={notificationSettings.webhookSecret}
                            onChange={val => handleNotificationSettingChange('webhookSecret', val)}
                            placeholder={t('请输入密钥')}
                          />
                          <Typography.Text type="secondary" style={{ marginTop: 8, display: 'block' }}>
                            {t('密钥将以 Bearer 方式添加到请求头中，用于验证webhook请求的合法性')}
                          </Typography.Text>
                          <Typography.Text type="secondary" style={{ marginTop: 4, display: 'block' }}>
                            {t('Authorization: Bearer your-secret-key')}
                          </Typography.Text>
                        </div>
                      </div>
                    </>
                  )}
                  {notificationSettings.warningType === 'email' && (
                    <div style={{ marginTop: 20 }}>
                      <Typography.Text strong>{t('通知邮箱')}</Typography.Text>
                      <div style={{ marginTop: 10 }}>
                        <Input
                          value={notificationSettings.notificationEmail}
                          onChange={val => handleNotificationSettingChange('notificationEmail', val)}
                          placeholder={t('留空则使用账号绑定的邮箱')}
                        />
                        <Typography.Text type="secondary" style={{ marginTop: 8, display: 'block' }}>
                          {t('设置用于接收额度预警的邮箱地址，不填则使用账号绑定的邮箱')}
                        </Typography.Text>
                      </div>
                    </div>
                  )}
                  <div style={{ marginTop: 20 }}>
                    <Typography.Text
                      strong>{t('额度预警阈值')} {renderQuotaWithPrompt(notificationSettings.warningThreshold)}</Typography.Text>
                    <div style={{ marginTop: 10 }}>
                      <AutoComplete
                        value={notificationSettings.warningThreshold}
                        onChange={val => handleNotificationSettingChange('warningThreshold', val)}
                        style={{ width: 200 }}
                        placeholder={t('请输入预警额度')}
                        data={[
                          { value: 100000, label: '0.2$' },
                          { value: 500000, label: '1$' },
                          { value: 1000000, label: '5$' },
                          { value: 5000000, label: '10$' }
                        ]}
                      />
                    </div>
                    <Typography.Text type="secondary" style={{ marginTop: 10, display: 'block' }}>
                      {t('当剩余额度低于此数值时，系统将通过选择的方式发送通知')}
                    </Typography.Text>
                  </div>
                </TabPane>
              </Tabs>
              <div style={{ marginTop: 20 }}>
                <Button type="primary" onClick={saveNotificationSettings}>
                  {t('保存设置')}
                </Button>
              </div>
            </Card>
            <Modal
              onCancel={() => setShowEmailBindModal(false)}
              onOk={bindEmail}
              visible={showEmailBindModal}
              size={'small'}
              centered={true}
              maskClosable={false}
            >
              <Typography.Title heading={6}>{t('绑定邮箱地址')}</Typography.Title>
              <div
                style={{
                  marginTop: 20,
                  display: 'flex',
                  justifyContent: 'space-between'
                }}
              >
                <Input
                  fluid
                  placeholder="输入邮箱地址"
                  onChange={(value) => handleInputChange('email', value)}
                  name="email"
                  type="email"
                />
                <Button
                  onClick={sendVerificationCode}
                  disabled={disableButton || loading}
                >
                  {disableButton ? `重新发送 (${countdown})` : '获取验证码'}
                </Button>
              </div>
              <div style={{ marginTop: 10 }}>
                <Input
                  fluid
                  placeholder="验证码"
                  name="email_verification_code"
                  value={inputs.email_verification_code}
                  onChange={(value) =>
                    handleInputChange('email_verification_code', value)
                  }
                />
              </div>
              {turnstileEnabled ? (
                <Turnstile
                  sitekey={turnstileSiteKey}
                  onVerify={(token) => {
                    setTurnstileToken(token);
                  }}
                />
              ) : (
                <></>
              )}
            </Modal>
            <Modal
              onCancel={() => setShowAccountDeleteModal(false)}
              visible={showAccountDeleteModal}
              size={'small'}
              centered={true}
              onOk={deleteAccount}
            >
              <div style={{ marginTop: 20 }}>
                <Banner
                  type="danger"
                  description="您正在删除自己的帐户，将清空所有数据且不可恢复"
                  closeIcon={null}
                />
              </div>
              <div style={{ marginTop: 20 }}>
                <Input
                  placeholder={`输入你的账户名 ${userState?.user?.username} 以确认删除`}
                  name="self_account_deletion_confirmation"
                  value={inputs.self_account_deletion_confirmation}
                  onChange={(value) =>
                    handleInputChange(
                      'self_account_deletion_confirmation',
                      value
                    )
                  }
                />
                {turnstileEnabled ? (
                  <Turnstile
                    sitekey={turnstileSiteKey}
                    onVerify={(token) => {
                      setTurnstileToken(token);
                    }}
                  />
                ) : (
                  <></>
                )}
              </div>
            </Modal>
            <Modal
              onCancel={() => setShowChangePasswordModal(false)}
              visible={showChangePasswordModal}
              size={'small'}
              centered={true}
              onOk={changePassword}
            >
              <div style={{ marginTop: 20 }}>
                <Input
                  name="set_new_password"
                  placeholder={t('新密码')}
                  value={inputs.set_new_password}
                  onChange={(value) =>
                    handleInputChange('set_new_password', value)
                  }
                />
                <Input
                  style={{ marginTop: 20 }}
                  name="set_new_password_confirmation"
                  placeholder={t('确认新密码')}
                  value={inputs.set_new_password_confirmation}
                  onChange={(value) =>
                    handleInputChange('set_new_password_confirmation', value)
                  }
                />
                {turnstileEnabled ? (
                  <Turnstile
                    sitekey={turnstileSiteKey}
                    onVerify={(token) => {
                      setTurnstileToken(token);
                    }}
                  />
                ) : (
                  <></>
                )}
              </div>
            </Modal>
          </div>
        </Layout.Content>
      </Layout>
    </div>
  );
};

export default PersonalSetting;
