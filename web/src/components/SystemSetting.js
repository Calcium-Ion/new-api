import React, { useEffect, useState, useRef } from 'react';
import {
  Button,
  Form,
  Row,
  Col,
  Typography,
  Modal,
  Banner,
  TagInput,
  Spin,
} from '@douyinfe/semi-ui';
const { Text } = Typography;
import {
  removeTrailingSlash,
  showError,
  showSuccess,
  verifyJSON,
} from '../helpers/utils';
import { API } from '../helpers/api';

const SystemSetting = () => {
  let [inputs, setInputs] = useState({
    PasswordLoginEnabled: '',
    PasswordRegisterEnabled: '',
    EmailVerificationEnabled: '',
    GitHubOAuthEnabled: '',
    GitHubClientId: '',
    GitHubClientSecret: '',
    'oidc.enabled': '',
    'oidc.client_id': '',
    'oidc.client_secret': '',
    'oidc.well_known': '',
    'oidc.authorization_endpoint': '',
    'oidc.token_endpoint': '',
    'oidc.user_info_endpoint': '',
    Notice: '',
    SMTPServer: '',
    SMTPPort: '',
    SMTPAccount: '',
    SMTPFrom: '',
    SMTPToken: '',
    ServerAddress: '',
    WorkerUrl: '',
    WorkerValidKey: '',
    EpayId: '',
    EpayKey: '',
    Price: 7.3,
    MinTopUp: 1,
    TopupGroupRatio: '',
    PayAddress: '',
    CustomCallbackAddress: '',
    Footer: '',
    WeChatAuthEnabled: '',
    WeChatServerAddress: '',
    WeChatServerToken: '',
    WeChatAccountQRCodeImageURL: '',
    TurnstileCheckEnabled: '',
    TurnstileSiteKey: '',
    TurnstileSecretKey: '',
    RegisterEnabled: '',
    EmailDomainRestrictionEnabled: '',
    EmailAliasRestrictionEnabled: '',
    SMTPSSLEnabled: '',
    EmailDomainWhitelist: [],
    // telegram login
    TelegramOAuthEnabled: '',
    TelegramBotToken: '',
    TelegramBotName: '',
    LinuxDOOAuthEnabled: '',
    LinuxDOClientId: '',
    LinuxDOClientSecret: '',
  });

  const [originInputs, setOriginInputs] = useState({});
  const [loading, setLoading] = useState(false);
  const [isLoaded, setIsLoaded] = useState(false);
  const formApiRef = useRef(null);
  const [emailDomainWhitelist, setEmailDomainWhitelist] = useState([]);
  const [showPasswordLoginConfirmModal, setShowPasswordLoginConfirmModal] = useState(false);
  const [linuxDOOAuthEnabled, setLinuxDOOAuthEnabled] = useState(false);

  const getOptions = async () => {
    setLoading(true);
    const res = await API.get('/api/option/');
    const { success, message, data } = res.data;
    if (success) {
      let newInputs = {};
      data.forEach((item) => {
        switch (item.key) {
          case 'TopupGroupRatio':
            item.value = JSON.stringify(JSON.parse(item.value), null, 2);
            break;
          case 'EmailDomainWhitelist':
            setEmailDomainWhitelist(item.value ? item.value.split(',') : []);
            break;
          case 'PasswordLoginEnabled':
          case 'PasswordRegisterEnabled':
          case 'EmailVerificationEnabled':
          case 'GitHubOAuthEnabled':
          case 'WeChatAuthEnabled':
          case 'TelegramOAuthEnabled':
          case 'RegisterEnabled':
          case 'TurnstileCheckEnabled':
          case 'EmailDomainRestrictionEnabled':
          case 'EmailAliasRestrictionEnabled':
          case 'SMTPSSLEnabled':
          case 'LinuxDOOAuthEnabled':
          case 'oidc.enabled':
            item.value = item.value === 'true';
            break;
          case 'Price':
          case 'MinTopUp':
            item.value = parseFloat(item.value);
            break;
          default:
            break;
        }
        newInputs[item.key] = item.value;
      });
      setInputs(newInputs);
      setOriginInputs(newInputs);
      if (formApiRef.current) {
        formApiRef.current.setValues(newInputs);
      }
      setIsLoaded(true);
    } else {
      showError(message);
    }
    setLoading(false);
  };

  useEffect(() => {
    getOptions();
  }, []);

  const updateOptions = async (options) => {
    setLoading(true);
    try {
      // 分离 checkbox 类型的选项和其他选项
      const checkboxOptions = options.filter(opt => 
        opt.key.toLowerCase().endsWith('enabled')
      );
      const otherOptions = options.filter(opt => 
        !opt.key.toLowerCase().endsWith('enabled')
      );

      // 处理 checkbox 类型的选项
      for (const opt of checkboxOptions) {
        const res = await API.put('/api/option/', {
          key: opt.key,
          value: opt.value.toString()
        });
        if (!res.data.success) {
          showError(res.data.message);
          return;
        }
      }

      // 处理其他选项
      if (otherOptions.length > 0) {
        const requestQueue = otherOptions.map(opt => 
          API.put('/api/option/', {
            key: opt.key,
            value: typeof opt.value === 'boolean' ? opt.value.toString() : opt.value
          })
        );

        const results = await Promise.all(requestQueue);
        
        // 检查所有请求是否成功
        const errorResults = results.filter(res => !res.data.success);
        errorResults.forEach(res => {
          showError(res.data.message);
        });
      }

      showSuccess('更新成功');
      // 更新本地状态
      const newInputs = { ...inputs };
      options.forEach(opt => {
        newInputs[opt.key] = opt.value;
      });
      setInputs(newInputs);
    } catch (error) {
      showError('更新失败');
    }
    setLoading(false);
  };

  const handleFormChange = (values) => {
    setInputs(values);
  };

  const submitServerAddress = async () => {
    let ServerAddress = removeTrailingSlash(inputs.ServerAddress);
    await updateOptions([{ key: 'ServerAddress', value: ServerAddress }]);
  };

  const submitWorker = async () => {
    let WorkerUrl = removeTrailingSlash(inputs.WorkerUrl);
    await updateOptions([
      { key: 'WorkerUrl', value: WorkerUrl },
      { key: 'WorkerValidKey', value: inputs.WorkerValidKey }
    ]);
  };

  const submitPayAddress = async () => {
    if (inputs.ServerAddress === '') {
      showError('请先填写服务器地址');
      return;
    }
    if (originInputs['TopupGroupRatio'] !== inputs.TopupGroupRatio) {
      if (!verifyJSON(inputs.TopupGroupRatio)) {
        showError('充值分组倍率不是合法的 JSON 字符串');
        return;
      }
    }
    
    const options = [
      { key: 'PayAddress', value: removeTrailingSlash(inputs.PayAddress) }
    ];
    
    if (inputs.EpayId !== '') {
      options.push({ key: 'EpayId', value: inputs.EpayId });
    }
    if (inputs.EpayKey !== undefined && inputs.EpayKey !== '') {
      options.push({ key: 'EpayKey', value: inputs.EpayKey });
    }
    if (inputs.Price !== '') {
      options.push({ key: 'Price', value: inputs.Price.toString() });
    }
    if (inputs.MinTopUp !== '') {
      options.push({ key: 'MinTopUp', value: inputs.MinTopUp.toString() });
    }
    if (inputs.CustomCallbackAddress !== '') {
      options.push({ key: 'CustomCallbackAddress', value: inputs.CustomCallbackAddress });
    }
    if (originInputs['TopupGroupRatio'] !== inputs.TopupGroupRatio) {
      options.push({ key: 'TopupGroupRatio', value: inputs.TopupGroupRatio });
    }
    
    await updateOptions(options);
  };

  const submitSMTP = async () => {
    const options = [];
    
    if (originInputs['SMTPServer'] !== inputs.SMTPServer) {
      options.push({ key: 'SMTPServer', value: inputs.SMTPServer });
    }
    if (originInputs['SMTPAccount'] !== inputs.SMTPAccount) {
      options.push({ key: 'SMTPAccount', value: inputs.SMTPAccount });
    }
    if (originInputs['SMTPFrom'] !== inputs.SMTPFrom) {
      options.push({ key: 'SMTPFrom', value: inputs.SMTPFrom });
    }
    if (originInputs['SMTPPort'] !== inputs.SMTPPort && inputs.SMTPPort !== '') {
      options.push({ key: 'SMTPPort', value: inputs.SMTPPort });
    }
    if (originInputs['SMTPToken'] !== inputs.SMTPToken && inputs.SMTPToken !== '') {
      options.push({ key: 'SMTPToken', value: inputs.SMTPToken });
    }
    
    if (options.length > 0) {
      await updateOptions(options);
    }
  };

  const submitEmailDomainWhitelist = async () => {
    if (Array.isArray(emailDomainWhitelist)) {
      await updateOptions([{ 
        key: 'EmailDomainWhitelist', 
        value: emailDomainWhitelist.join(',') 
      }]);
    } else {
      showError('邮箱域名白名单格式不正确');
    }
  };

  const submitWeChat = async () => {
    const options = [];
    
    if (originInputs['WeChatServerAddress'] !== inputs.WeChatServerAddress) {
      options.push({ 
        key: 'WeChatServerAddress', 
        value: removeTrailingSlash(inputs.WeChatServerAddress) 
      });
    }
    if (originInputs['WeChatAccountQRCodeImageURL'] !== inputs.WeChatAccountQRCodeImageURL) {
      options.push({ 
        key: 'WeChatAccountQRCodeImageURL', 
        value: inputs.WeChatAccountQRCodeImageURL 
      });
    }
    if (originInputs['WeChatServerToken'] !== inputs.WeChatServerToken && inputs.WeChatServerToken !== '') {
      options.push({ key: 'WeChatServerToken', value: inputs.WeChatServerToken });
    }
    
    if (options.length > 0) {
      await updateOptions(options);
    }
  };

  const submitGitHubOAuth = async () => {
    const options = [];
    
    if (originInputs['GitHubClientId'] !== inputs.GitHubClientId) {
      options.push({ key: 'GitHubClientId', value: inputs.GitHubClientId });
    }
    if (originInputs['GitHubClientSecret'] !== inputs.GitHubClientSecret && inputs.GitHubClientSecret !== '') {
      options.push({ key: 'GitHubClientSecret', value: inputs.GitHubClientSecret });
    }
    
    if (options.length > 0) {
      await updateOptions(options);
    }
  };

  const submitOIDCSettings = async () => {
    if (inputs['oidc.well_known'] !== '') {
      if (!inputs['oidc.well_known'].startsWith('http://') && !inputs['oidc.well_known'].startsWith('https://')) {
        showError('Well-Known URL 必须以 http:// 或 https:// 开头');
        return;
      }
      try {
        const res = await API.get(inputs['oidc.well_known']);
        inputs['oidc.authorization_endpoint'] = res.data['authorization_endpoint'];
        inputs['oidc.token_endpoint'] = res.data['token_endpoint'];
        inputs['oidc.user_info_endpoint'] = res.data['userinfo_endpoint'];
        showSuccess('获取 OIDC 配置成功！');
      } catch (err) {
        console.error(err);
        showError("获取 OIDC 配置失败，请检查网络状况和 Well-Known URL 是否正确");
        return;
      }
    }

    const options = [];
    
    if (originInputs['oidc.well_known'] !== inputs['oidc.well_known']) {
      options.push({ key: 'oidc.well_known', value: inputs['oidc.well_known'] });
    }
    if (originInputs['oidc.client_id'] !== inputs['oidc.client_id']) {
      options.push({ key: 'oidc.client_id', value: inputs['oidc.client_id'] });
    }
    if (originInputs['oidc.client_secret'] !== inputs['oidc.client_secret'] && inputs['oidc.client_secret'] !== '') {
      options.push({ key: 'oidc.client_secret', value: inputs['oidc.client_secret'] });
    }
    if (originInputs['oidc.authorization_endpoint'] !== inputs['oidc.authorization_endpoint']) {
      options.push({ key: 'oidc.authorization_endpoint', value: inputs['oidc.authorization_endpoint'] });
    }
    if (originInputs['oidc.token_endpoint'] !== inputs['oidc.token_endpoint']) {
      options.push({ key: 'oidc.token_endpoint', value: inputs['oidc.token_endpoint'] });
    }
    if (originInputs['oidc.user_info_endpoint'] !== inputs['oidc.user_info_endpoint']) {
      options.push({ key: 'oidc.user_info_endpoint', value: inputs['oidc.user_info_endpoint'] });
    }
    
    if (options.length > 0) {
      await updateOptions(options);
    }
  };

  const submitTelegramSettings = async () => {
    const options = [
      { key: 'TelegramBotToken', value: inputs.TelegramBotToken },
      { key: 'TelegramBotName', value: inputs.TelegramBotName }
    ];
    await updateOptions(options);
  };

  const submitTurnstile = async () => {
    const options = [];
    
    if (originInputs['TurnstileSiteKey'] !== inputs.TurnstileSiteKey) {
      options.push({ key: 'TurnstileSiteKey', value: inputs.TurnstileSiteKey });
    }
    if (originInputs['TurnstileSecretKey'] !== inputs.TurnstileSecretKey && inputs.TurnstileSecretKey !== '') {
      options.push({ key: 'TurnstileSecretKey', value: inputs.TurnstileSecretKey });
    }
    
    if (options.length > 0) {
      await updateOptions(options);
    }
  };

  const submitLinuxDOOAuth = async () => {
    const options = [];
    
    if (originInputs['LinuxDOClientId'] !== inputs.LinuxDOClientId) {
      options.push({ key: 'LinuxDOClientId', value: inputs.LinuxDOClientId });
    }
    if (originInputs['LinuxDOClientSecret'] !== inputs.LinuxDOClientSecret && inputs.LinuxDOClientSecret !== '') {
      options.push({ key: 'LinuxDOClientSecret', value: inputs.LinuxDOClientSecret });
    }
    
    if (options.length > 0) {
      await updateOptions(options);
    }
  };

  const handleCheckboxChange = async (optionKey, event) => {
    const value = event.target.checked;
    
    if (optionKey === 'PasswordLoginEnabled' && !value) {
      setShowPasswordLoginConfirmModal(true);
    } else {
      await updateOptions([{ key: optionKey, value }]);
    }
    if (optionKey === 'LinuxDOOAuthEnabled') {
      setLinuxDOOAuthEnabled(value);
    }
  };

  const handlePasswordLoginConfirm = async () => {
    await updateOptions([{ key: 'PasswordLoginEnabled', value: false }]);
    setShowPasswordLoginConfirmModal(false);
  };

  return (
    <div style={{ padding: '20px' }}>
      {isLoaded ? (
        <Form
          initValues={inputs}
          onValueChange={handleFormChange}
          getFormApi={(api) => (formApiRef.current = api)}
        >
          {({ formState, values, formApi }) => (
            <>
              <Form.Section text='通用设置'>
                <Form.Input
                  field='ServerAddress'
                  label='服务器地址'
                  placeholder='例如：https://yourdomain.com'
                  style={{ width: '100%' }}
                />
                <Button onClick={submitServerAddress}>更新服务器地址</Button>
              </Form.Section>

              <Form.Section text='代理设置'>
                <Text>
                  （支持{' '}
                  <a
                    href='https://github.com/Calcium-Ion/new-api-worker'
                    target='_blank'
                    rel='noreferrer'
                  >
                    new-api-worker
                  </a>
                  ）
                </Text>
                <Row gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}>
                  <Col xs={24} sm={24} md={12} lg={12} xl={12}>
                    <Form.Input
                      field='WorkerUrl'
                      label='Worker地址'
                      placeholder='例如：https://workername.yourdomain.workers.dev'
                    />
                  </Col>
                  <Col xs={24} sm={24} md={12} lg={12} xl={12}>
                    <Form.Input
                      field='WorkerValidKey'
                      label='Worker密钥'
                      placeholder='敏感信息不会发送到前端显示'
                      type='password'
                    />
                  </Col>
                </Row>
                <Button onClick={submitWorker}>更新Worker设置</Button>
              </Form.Section>

              <Form.Section text='支付设置'>
                <Text>
                  （当前仅支持易支付接口，默认使用上方服务器地址作为回调地址！）
                </Text>
                <Row gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}>
                  <Col xs={24} sm={24} md={8} lg={8} xl={8}>
                    <Form.Input
                      field='PayAddress'
                      label='支付地址'
                      placeholder='例如：https://yourdomain.com'
                    />
                  </Col>
                  <Col xs={24} sm={24} md={8} lg={8} xl={8}>
                    <Form.Input
                      field='EpayId'
                      label='易支付商户ID'
                      placeholder='例如：0001'
                    />
                  </Col>
                  <Col xs={24} sm={24} md={8} lg={8} xl={8}>
                    <Form.Input
                      field='EpayKey'
                      label='易支付商户密钥'
                      placeholder='敏感信息不会发送到前端显示'
                      type='password'
                    />
                  </Col>
                </Row>
                <Row
                  gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}
                  style={{ marginTop: 16 }}
                >
                  <Col xs={24} sm={24} md={8} lg={8} xl={8}>
                    <Form.Input
                      field='CustomCallbackAddress'
                      label='回调地址'
                      placeholder='例如：https://yourdomain.com'
                    />
                  </Col>
                  <Col xs={24} sm={24} md={8} lg={8} xl={8}>
                    <Form.InputNumber
                      field='Price'
                      precision={2}
                      label='充值价格（x元/美金）'
                      placeholder='例如：7，就是7元/美金'
                    />
                  </Col>
                  <Col xs={24} sm={24} md={8} lg={8} xl={8}>
                    <Form.InputNumber
                      field='MinTopUp'
                      label='最低充值美元数量'
                      placeholder='例如：2，就是最低充值2$'
                    />
                  </Col>
                </Row>
                <Form.TextArea
                  field='TopupGroupRatio'
                  label='充值分组倍率'
                  placeholder='为一个 JSON 文本，键为组名称，值为倍率'
                  autosize
                />
                <Button onClick={submitPayAddress}>更新支付设置</Button>
              </Form.Section>

              <Form.Section text='配置登录注册'>
                <Row gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}>
                  <Col xs={24} sm={24} md={12} lg={12} xl={12}>
                    <Form.Checkbox
                      field='PasswordLoginEnabled'
                      noLabel
                      onChange={(e) =>
                        handleCheckboxChange('PasswordLoginEnabled', e)
                      }
                    >
                      允许通过密码进行登录
                    </Form.Checkbox>
                    <Form.Checkbox
                      field='PasswordRegisterEnabled'
                      noLabel
                      onChange={(e) =>
                        handleCheckboxChange('PasswordRegisterEnabled', e)
                      }
                    >
                      允许通过密码进行注册
                    </Form.Checkbox>
                    <Form.Checkbox
                      field='EmailVerificationEnabled'
                      noLabel
                      onChange={(e) =>
                        handleCheckboxChange('EmailVerificationEnabled', e)
                      }
                    >
                      通过密码注册时需要进行邮箱验证
                    </Form.Checkbox>
                    <Form.Checkbox
                      field='RegisterEnabled'
                      noLabel
                      onChange={(e) => handleCheckboxChange('RegisterEnabled', e)}
                    >
                      允许新用户注册
                    </Form.Checkbox>
                    <Form.Checkbox
                      field='TurnstileCheckEnabled'
                      noLabel
                      onChange={(e) =>
                        handleCheckboxChange('TurnstileCheckEnabled', e)
                      }
                    >
                      启用 Turnstile 用户校验
                    </Form.Checkbox>
                  </Col>
                  <Col xs={24} sm={24} md={12} lg={12} xl={12}>
                    <Form.Checkbox
                      field='GitHubOAuthEnabled'
                      noLabel
                      onChange={(e) =>
                        handleCheckboxChange('GitHubOAuthEnabled', e)
                      }
                    >
                      允许通过 GitHub 账户登录 & 注册
                    </Form.Checkbox>
                    <Form.Checkbox
                      field='LinuxDOOAuthEnabled'
                      noLabel
                      onChange={(e) =>
                        handleCheckboxChange('LinuxDOOAuthEnabled', e)
                      }
                    >
                      允许通过 Linux DO 账户登录 & 注册
                    </Form.Checkbox>
                    <Form.Checkbox
                      field='WeChatAuthEnabled'
                      noLabel
                      onChange={(e) =>
                        handleCheckboxChange('WeChatAuthEnabled', e)
                      }
                    >
                      允许通过微信登录 & 注册
                    </Form.Checkbox>
                    <Form.Checkbox
                      field='TelegramOAuthEnabled'
                      noLabel
                      onChange={(e) =>
                        handleCheckboxChange('TelegramOAuthEnabled', e)
                      }
                    >
                      允许通过 Telegram 进行登录
                    </Form.Checkbox>
                    <Form.Checkbox
                      field='oidc.enabled'
                      noLabel
                      onChange={(e) => handleCheckboxChange('oidc.enabled', e)}
                    >
                      允许通过 OIDC 进行登录
                    </Form.Checkbox>
                  </Col>
                </Row>
              </Form.Section>

              <Form.Section text='配置邮箱域名白名单'>
                <Text>用以防止恶意用户利用临时邮箱批量注册</Text>
                <Row gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}>
                  <Col xs={24} sm={24} md={12} lg={12} xl={12}>
                    <Form.Checkbox
                      field='EmailDomainRestrictionEnabled'
                      noLabel
                      onChange={(e) =>
                        handleCheckboxChange('EmailDomainRestrictionEnabled', e)
                      }
                    >
                      启用邮箱域名白名单
                    </Form.Checkbox>
                  </Col>
                  <Col xs={24} sm={24} md={12} lg={12} xl={12}>
                    <Form.Checkbox
                      field='EmailAliasRestrictionEnabled'
                      noLabel
                      onChange={(e) =>
                        handleCheckboxChange('EmailAliasRestrictionEnabled', e)
                      }
                    >
                      启用邮箱别名限制
                    </Form.Checkbox>
                  </Col>
                </Row>
                <TagInput
                  value={emailDomainWhitelist}
                  onChange={setEmailDomainWhitelist}
                  placeholder='输入域名后回车'
                  style={{ width: '100%', marginTop: 16 }}
                />
                <Button
                  onClick={submitEmailDomainWhitelist}
                  style={{ marginTop: 10 }}
                >
                  保存邮箱域名白名单设置
                </Button>
              </Form.Section>

              <Form.Section text='配置 SMTP'>
                <Text>用以支持系统的邮件发送</Text>
                <Row gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}>
                  <Col xs={24} sm={24} md={8} lg={8} xl={8}>
                    <Form.Input field='SMTPServer' label='SMTP 服务器地址' />
                  </Col>
                  <Col xs={24} sm={24} md={8} lg={8} xl={8}>
                    <Form.Input field='SMTPPort' label='SMTP 端口' />
                  </Col>
                  <Col xs={24} sm={24} md={8} lg={8} xl={8}>
                    <Form.Input field='SMTPAccount' label='SMTP 账户' />
                  </Col>
                </Row>
                <Row
                  gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}
                  style={{ marginTop: 16 }}
                >
                  <Col xs={24} sm={24} md={8} lg={8} xl={8}>
                    <Form.Input field='SMTPFrom' label='SMTP 发送者邮箱' />
                  </Col>
                  <Col xs={24} sm={24} md={8} lg={8} xl={8}>
                    <Form.Input
                      field='SMTPToken'
                      label='SMTP 访问凭证'
                      type='password'
                      placeholder='敏感信息不会发送到前端显示'
                    />
                  </Col>
                  <Col xs={24} sm={24} md={8} lg={8} xl={8}>
                    <Form.Checkbox
                      field='SMTPSSLEnabled'
                      noLabel
                      onChange={(e) => handleCheckboxChange('SMTPSSLEnabled', e)}
                    >
                      启用SMTP SSL
                    </Form.Checkbox>
                  </Col>
                </Row>
                <Button onClick={submitSMTP}>保存 SMTP 设置</Button>
              </Form.Section>

              <Form.Section text='配置 OIDC'>
                <Text>用以支持通过 OIDC 登录，例如 Okta、Auth0 等兼容 OIDC 协议的 IdP</Text>
                <Banner
                  type='info'
                  description={`主页链接填 ${inputs.ServerAddress ? inputs.ServerAddress : '网站地址'}，重定向 URL 填 ${inputs.ServerAddress ? inputs.ServerAddress : '网站地址'}/oauth/oidc`}
                  style={{ marginBottom: 20, marginTop: 16 }}
                />
                <Text>若你的 OIDC Provider 支持 Discovery Endpoint，你可以仅填写 OIDC Well-Known URL，系统会自动获取 OIDC 配置</Text>
                <Row gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}>
                  <Col xs={24} sm={24} md={12} lg={12} xl={12}>
                    <Form.Input
                      field='oidc.well_known'
                      label='Well-Known URL'
                      placeholder='请输入 OIDC 的 Well-Known URL'
                    />
                  </Col>
                  <Col xs={24} sm={24} md={12} lg={12} xl={12}>
                    <Form.Input
                      field='oidc.client_id'
                      label='Client ID'
                      placeholder='输入 OIDC 的 Client ID'
                    />
                  </Col>
                </Row>
                <Row gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}>
                  <Col xs={24} sm={24} md={12} lg={12} xl={12}>
                    <Form.Input
                      field='oidc.client_secret'
                      label='Client Secret'
                      type='password'
                      placeholder='敏感信息不会发送到前端显示'
                    />
                  </Col>
                  <Col xs={24} sm={24} md={12} lg={12} xl={12}>
                    <Form.Input
                      field='oidc.authorization_endpoint'
                      label='Authorization Endpoint'
                      placeholder='输入 OIDC 的 Authorization Endpoint'
                    />
                  </Col>
                </Row>
                <Row gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}>
                  <Col xs={24} sm={24} md={12} lg={12} xl={12}>
                    <Form.Input
                      field='oidc.token_endpoint'
                      label='Token Endpoint'
                      placeholder='输入 OIDC 的 Token Endpoint'
                    />
                  </Col>
                  <Col xs={24} sm={24} md={12} lg={12} xl={12}>
                    <Form.Input
                      field='oidc.user_info_endpoint'
                      label='User Info Endpoint'
                      placeholder='输入 OIDC 的 Userinfo Endpoint'
                    />
                  </Col>
                </Row>
                <Button onClick={submitOIDCSettings}>保存 OIDC 设置</Button>
              </Form.Section>

              <Form.Section text='配置 GitHub OAuth App'>
                <Text>用以支持通过 GitHub 进行登录注册</Text>
                <Banner
                  type='info'
                  description={`Homepage URL 填 ${inputs.ServerAddress ? inputs.ServerAddress : '网站地址'}，Authorization callback URL 填 ${inputs.ServerAddress ? inputs.ServerAddress : '网站地址'}/oauth/github`}
                  style={{ marginBottom: 20, marginTop: 16 }}
                />
                <Row gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}>
                  <Col xs={24} sm={24} md={12} lg={12} xl={12}>
                    <Form.Input field='GitHubClientId' label='GitHub Client ID' />
                  </Col>
                  <Col xs={24} sm={24} md={12} lg={12} xl={12}>
                    <Form.Input
                      field='GitHubClientSecret'
                      label='GitHub Client Secret'
                      type='password'
                      placeholder='敏感信息不会发送到前端显示'
                    />
                  </Col>
                </Row>
                <Button onClick={submitGitHubOAuth}>
                  保存 GitHub OAuth 设置
                </Button>
              </Form.Section>
              <Form.Section text='配置 Linux DO OAuth'>
                <Text>
                  用以支持通过 Linux DO 进行登录注册
                  <a
                    href='https://connect.linux.do/'
                    target='_blank'
                    rel='noreferrer'
                    style={{ display: 'inline-block', marginLeft: 4, marginRight: 4 }}
                  >
                    点击此处
                  </a>
                  管理你的 LinuxDO OAuth App
                </Text>
                <Banner
                  type='info'
                  description={`回调 URL 填 ${inputs.ServerAddress ? inputs.ServerAddress : '网站地址'}/oauth/linuxdo`}
                  style={{ marginBottom: 20, marginTop: 16 }}
                />
                <Row gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}>
                  <Col xs={24} sm={24} md={12} lg={12} xl={12}>
                    <Form.Input 
                      field='LinuxDOClientId' 
                      label='Linux DO Client ID'
                      placeholder='输入你注册的 LinuxDO OAuth APP 的 ID'
                    />
                  </Col>
                  <Col xs={24} sm={24} md={12} lg={12} xl={12}>
                    <Form.Input
                      field='LinuxDOClientSecret'
                      label='Linux DO Client Secret'
                      type='password'
                      placeholder='敏感信息不会发送到前端显示'
                    />
                  </Col>
                </Row>
                <Button onClick={submitLinuxDOOAuth}>
                  保存 Linux DO OAuth 设置
                </Button>
              </Form.Section>
              <Form.Section text='配置 WeChat Server'>
                <Text>用以支持通过微信进行登录注册</Text>
                <Row gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}>
                  <Col xs={24} sm={24} md={8} lg={8} xl={8}>
                    <Form.Input
                      field='WeChatServerAddress'
                      label='WeChat Server 服务器地址'
                    />
                  </Col>
                  <Col xs={24} sm={24} md={8} lg={8} xl={8}>
                    <Form.Input
                      field='WeChatServerToken'
                      label='WeChat Server 访问凭证'
                      type='password'
                      placeholder='敏感信息不会发送到前端显示'
                    />
                  </Col>
                  <Col xs={24} sm={24} md={8} lg={8} xl={8}>
                    <Form.Input
                      field='WeChatAccountQRCodeImageURL'
                      label='微信公众号二维码图片链接'
                    />
                  </Col>
                </Row>
                <Button onClick={submitWeChat}>保存 WeChat Server 设置</Button>
              </Form.Section>
              <Form.Section text='配置 Telegram 登录'>
                <Text>用以支持通过 Telegram 进行登录注册</Text>
                <Row gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}>
                  <Col xs={24} sm={24} md={12} lg={12} xl={12}>
                    <Form.Input
                      field='TelegramBotToken'
                      label='Telegram Bot Token'
                      placeholder='敏感信息不会发送到前端显示'
                      type='password'
                    />
                  </Col>
                  <Col xs={24} sm={24} md={12} lg={12} xl={12}>
                    <Form.Input
                      field='TelegramBotName'
                      label='Telegram Bot 名称'
                    />
                  </Col>
                </Row>
                <Button onClick={submitTelegramSettings}>
                  保存 Telegram 登录设置
                </Button>
              </Form.Section>
              <Form.Section text='配置 Turnstile'>
                <Text>用以支持用户校验</Text>
                <Row gutter={{ xs: 8, sm: 16, md: 24, lg: 24, xl: 24, xxl: 24 }}>
                  <Col xs={24} sm={24} md={12} lg={12} xl={12}>
                    <Form.Input
                      field='TurnstileSiteKey'
                      label='Turnstile Site Key'
                    />
                  </Col>
                  <Col xs={24} sm={24} md={12} lg={12} xl={12}>
                    <Form.Input
                      field='TurnstileSecretKey'
                      label='Turnstile Secret Key'
                      type='password'
                      placeholder='敏感信息不会发送到前端显示'
                    />
                  </Col>
                </Row>
                <Button onClick={submitTurnstile}>保存 Turnstile 设置</Button>
              </Form.Section>
            
              <Modal
                title="确认取消密码登录"
                visible={showPasswordLoginConfirmModal}
                onOk={handlePasswordLoginConfirm}
                onCancel={() => {
                  setShowPasswordLoginConfirmModal(false);
                  formApiRef.current.setValue('PasswordLoginEnabled', true);
                }}
                okText="确认"
                cancelText="取消"
              >
                <p>您确定要取消密码登录功能吗？这可能会影响用户的登录方式。</p>
              </Modal>
            </>
          )}
        </Form>
      ) : (
        <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '100vh' }}>
          <Spin size="large" />
        </div>
      )}
    </div>
  );
};

export default SystemSetting;