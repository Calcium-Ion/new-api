import React, { useEffect, useState } from 'react';
import {
  Button,
  Divider,
  Form,
  Grid,
  Header,
  Message,
  Modal,
} from 'semantic-ui-react';
import { API, removeTrailingSlash, showError, verifyJSON } from '../helpers';

import { useTheme } from '../context/Theme';

const SystemSetting = () => {
  let [inputs, setInputs] = useState({
    PasswordLoginEnabled: '',
    PasswordRegisterEnabled: '',
    EmailVerificationEnabled: '',
    GitHubOAuthEnabled: '',
    GitHubClientId: '',
    GitHubClientSecret: '',
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
  let [loading, setLoading] = useState(false);
  const [EmailDomainWhitelist, setEmailDomainWhitelist] = useState([]);
  const [restrictedDomainInput, setRestrictedDomainInput] = useState('');
  const [showPasswordWarningModal, setShowPasswordWarningModal] =
    useState(false);

  const theme = useTheme();
  const isDark = theme === 'dark';

  const getOptions = async () => {
    const res = await API.get('/api/option/');
    const { success, message, data } = res.data;
    if (success) {
      let newInputs = {};
      data.forEach((item) => {
        if (item.key === 'TopupGroupRatio') {
          item.value = JSON.stringify(JSON.parse(item.value), null, 2);
        }
        newInputs[item.key] = item.value;
      });
      setInputs({
        ...newInputs,
        EmailDomainWhitelist: newInputs.EmailDomainWhitelist.split(','),
      });
      setOriginInputs(newInputs);

      setEmailDomainWhitelist(
        newInputs.EmailDomainWhitelist.split(',').map((item) => {
          return { key: item, text: item, value: item };
        }),
      );
    } else {
      showError(message);
    }
  };

  useEffect(() => {
    getOptions().then();
  }, []);
  useEffect(() => {}, [inputs.EmailDomainWhitelist]);

  const updateOption = async (key, value) => {
    setLoading(true);
    switch (key) {
      case 'PasswordLoginEnabled':
      case 'PasswordRegisterEnabled':
      case 'EmailVerificationEnabled':
      case 'GitHubOAuthEnabled':
      case 'LinuxDOOAuthEnabled':
      case 'WeChatAuthEnabled':
      case 'TelegramOAuthEnabled':
      case 'TurnstileCheckEnabled':
      case 'EmailDomainRestrictionEnabled':
      case 'EmailAliasRestrictionEnabled':
      case 'SMTPSSLEnabled':
      case 'RegisterEnabled':
        value = inputs[key] === 'true' ? 'false' : 'true';
        break;
      default:
        break;
    }
    const res = await API.put('/api/option/', {
      key,
      value,
    });
    const { success, message } = res.data;
    if (success) {
      if (key === 'EmailDomainWhitelist') {
        value = value.split(',');
      }
      if (key === 'Price') {
        value = parseFloat(value);
      }
      setInputs((inputs) => ({
        ...inputs,
        [key]: value,
      }));
    } else {
      showError(message);
    }
    setLoading(false);
  };

  const handleInputChange = async (e, { name, value }) => {
    if (name === 'PasswordLoginEnabled' && inputs[name] === 'true') {
      // block disabling password login
      setShowPasswordWarningModal(true);
      return;
    }
    if (
      name === 'Notice' ||
      (name.startsWith('SMTP') && name !== 'SMTPSSLEnabled') ||
      name === 'ServerAddress' ||
      name === 'WorkerUrl' ||
      name === 'WorkerValidKey' ||
      name === 'EpayId' ||
      name === 'EpayKey' ||
      name === 'Price' ||
      name === 'PayAddress' ||
      name === 'GitHubClientId' ||
      name === 'GitHubClientSecret' ||
      name === 'WeChatServerAddress' ||
      name === 'WeChatServerToken' ||
      name === 'WeChatAccountQRCodeImageURL' ||
      name === 'TurnstileSiteKey' ||
      name === 'TurnstileSecretKey' ||
      name === 'EmailDomainWhitelist' ||
      name === 'TopupGroupRatio' ||
      name === 'TelegramBotToken' ||
      name === 'TelegramBotName' ||
      name === 'LinuxDOClientId' ||
      name === 'LinuxDOClientSecret'
    ) {
      setInputs((inputs) => ({ ...inputs, [name]: value }));
    } else {
      await updateOption(name, value);
    }
  };

  const submitServerAddress = async () => {
    let ServerAddress = removeTrailingSlash(inputs.ServerAddress);
    await updateOption('ServerAddress', ServerAddress);
  };

  const submitWorker = async () => {
    let WorkerUrl = removeTrailingSlash(inputs.WorkerUrl);
    await updateOption('WorkerUrl', WorkerUrl);
    if (inputs.WorkerValidKey !== '') {
      await updateOption('WorkerValidKey', inputs.WorkerValidKey);
    }
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
      await updateOption('TopupGroupRatio', inputs.TopupGroupRatio);
    }
    let PayAddress = removeTrailingSlash(inputs.PayAddress);
    await updateOption('PayAddress', PayAddress);
    if (inputs.EpayId !== '') {
      await updateOption('EpayId', inputs.EpayId);
    }
    if (inputs.EpayKey !== undefined && inputs.EpayKey !== '') {
      await updateOption('EpayKey', inputs.EpayKey);
    }
    await updateOption('Price', '' + inputs.Price);
  };

  const submitSMTP = async () => {
    if (originInputs['SMTPServer'] !== inputs.SMTPServer) {
      await updateOption('SMTPServer', inputs.SMTPServer);
    }
    if (originInputs['SMTPAccount'] !== inputs.SMTPAccount) {
      await updateOption('SMTPAccount', inputs.SMTPAccount);
    }
    if (originInputs['SMTPFrom'] !== inputs.SMTPFrom) {
      await updateOption('SMTPFrom', inputs.SMTPFrom);
    }
    if (
      originInputs['SMTPPort'] !== inputs.SMTPPort &&
      inputs.SMTPPort !== ''
    ) {
      await updateOption('SMTPPort', inputs.SMTPPort);
    }
    if (
      originInputs['SMTPToken'] !== inputs.SMTPToken &&
      inputs.SMTPToken !== ''
    ) {
      await updateOption('SMTPToken', inputs.SMTPToken);
    }
  };

  const submitEmailDomainWhitelist = async () => {
    if (
      originInputs['EmailDomainWhitelist'] !==
        inputs.EmailDomainWhitelist.join(',') &&
      inputs.SMTPToken !== ''
    ) {
      await updateOption(
        'EmailDomainWhitelist',
        inputs.EmailDomainWhitelist.join(','),
      );
    }
  };

  const submitWeChat = async () => {
    if (originInputs['WeChatServerAddress'] !== inputs.WeChatServerAddress) {
      await updateOption(
        'WeChatServerAddress',
        removeTrailingSlash(inputs.WeChatServerAddress),
      );
    }
    if (
      originInputs['WeChatAccountQRCodeImageURL'] !==
      inputs.WeChatAccountQRCodeImageURL
    ) {
      await updateOption(
        'WeChatAccountQRCodeImageURL',
        inputs.WeChatAccountQRCodeImageURL,
      );
    }
    if (
      originInputs['WeChatServerToken'] !== inputs.WeChatServerToken &&
      inputs.WeChatServerToken !== ''
    ) {
      await updateOption('WeChatServerToken', inputs.WeChatServerToken);
    }
  };

  const submitGitHubOAuth = async () => {
    if (originInputs['GitHubClientId'] !== inputs.GitHubClientId) {
      await updateOption('GitHubClientId', inputs.GitHubClientId);
    }
    if (
      originInputs['GitHubClientSecret'] !== inputs.GitHubClientSecret &&
      inputs.GitHubClientSecret !== ''
    ) {
      await updateOption('GitHubClientSecret', inputs.GitHubClientSecret);
    }
  };

  const submitTelegramSettings = async () => {
    // await updateOption('TelegramOAuthEnabled', inputs.TelegramOAuthEnabled);
    await updateOption('TelegramBotToken', inputs.TelegramBotToken);
    await updateOption('TelegramBotName', inputs.TelegramBotName);
  };

  const submitTurnstile = async () => {
    if (originInputs['TurnstileSiteKey'] !== inputs.TurnstileSiteKey) {
      await updateOption('TurnstileSiteKey', inputs.TurnstileSiteKey);
    }
    if (
      originInputs['TurnstileSecretKey'] !== inputs.TurnstileSecretKey &&
      inputs.TurnstileSecretKey !== ''
    ) {
      await updateOption('TurnstileSecretKey', inputs.TurnstileSecretKey);
    }
  };

  const submitNewRestrictedDomain = () => {
    const localDomainList = inputs.EmailDomainWhitelist;
    if (
      restrictedDomainInput !== '' &&
      !localDomainList.includes(restrictedDomainInput)
    ) {
      setRestrictedDomainInput('');
      setInputs({
        ...inputs,
        EmailDomainWhitelist: [...localDomainList, restrictedDomainInput],
      });
      setEmailDomainWhitelist([
        ...EmailDomainWhitelist,
        {
          key: restrictedDomainInput,
          text: restrictedDomainInput,
          value: restrictedDomainInput,
        },
      ]);
    }
  };

  const submitLinuxDOOAuth = async () => {
    if (originInputs['LinuxDOClientId'] !== inputs.LinuxDOClientId) {
      await updateOption('LinuxDOClientId', inputs.LinuxDOClientId);
    }
    if (
      originInputs['LinuxDOClientSecret'] !== inputs.LinuxDOClientSecret &&
      inputs.LinuxDOClientSecret !== ''
    ) {
      await updateOption('LinuxDOClientSecret', inputs.LinuxDOClientSecret);
    }
  };

  return (
    <Grid columns={1}>
      <Grid.Column>
        <Form loading={loading} inverted={isDark}>
          <Header as='h3' inverted={isDark}>
            通用设置
          </Header>
          <Form.Group widths='equal'>
            <Form.Input
              label='服务器地址'
              placeholder='例如：https://yourdomain.com'
              value={inputs.ServerAddress}
              name='ServerAddress'
              onChange={handleInputChange}
            />
          </Form.Group>
          <Form.Button onClick={submitServerAddress}>
            更新服务器地址
          </Form.Button>
          <Header as='h3' inverted={isDark}>
            代理设置（支持{' '}
            <a
              href='https://github.com/Calcium-Ion/new-api-worker'
              target='_blank'
              rel='noreferrer'
            >
              new-api-worker
            </a>
            ）
          </Header>
          <Form.Group widths='equal'>
            <Form.Input
              label='Worker地址，不填写则不启用代理'
              placeholder='例如：https://workername.yourdomain.workers.dev'
              value={inputs.WorkerUrl}
              name='WorkerUrl'
              onChange={handleInputChange}
            />
            <Form.Input
              label='Worker密钥，根据你部署的 Worker 填写'
              placeholder='例如：your_secret_key'
              value={inputs.WorkerValidKey}
              name='WorkerValidKey'
              onChange={handleInputChange}
            />
          </Form.Group>
          <Form.Button onClick={submitWorker}>更新Worker设置</Form.Button>
          <Divider />
          <Header as='h3' inverted={isDark}>
            支付设置（当前仅支持易支付接口，默认使用上方服务器地址作为回调地址！）
          </Header>
          <Form.Group widths='equal'>
            <Form.Input
              label='支付地址，不填写则不启用在线支付'
              placeholder='例如：https://yourdomain.com'
              value={inputs.PayAddress}
              name='PayAddress'
              onChange={handleInputChange}
            />
            <Form.Input
              label='易支付商户ID'
              placeholder='例如：0001'
              value={inputs.EpayId}
              name='EpayId'
              onChange={handleInputChange}
            />
            <Form.Input
              label='易支付商户密钥'
              placeholder='敏感信息不会发送到前端显示'
              value={inputs.EpayKey}
              name='EpayKey'
              onChange={handleInputChange}
            />
          </Form.Group>
          <Form.Group widths='equal'>
            <Form.Input
              label='回调地址，不填写则使用上方服务器地址作为回调地址'
              placeholder='例如：https://yourdomain.com'
              value={inputs.CustomCallbackAddress}
              name='CustomCallbackAddress'
              onChange={handleInputChange}
            />
            <Form.Input
              label='充值价格（x元/美金）'
              placeholder='例如：7，就是7元/美金'
              value={inputs.Price}
              name='Price'
              min={0}
              onChange={handleInputChange}
            />
            <Form.Input
              label='最低充值美元数量（以美金为单位，如果使用额度请自行换算！）'
              placeholder='例如：2，就是最低充值2$'
              value={inputs.MinTopUp}
              name='MinTopUp'
              min={1}
              onChange={handleInputChange}
            />
          </Form.Group>
          <Form.Group widths='equal'>
            <Form.TextArea
              label='充值分组倍率'
              name='TopupGroupRatio'
              onChange={handleInputChange}
              style={{ minHeight: 250, fontFamily: 'JetBrains Mono, Consolas' }}
              autoComplete='new-password'
              value={inputs.TopupGroupRatio}
              placeholder='为一个 JSON 文本，键为组名称，值为倍率'
            />
          </Form.Group>
          <Form.Button onClick={submitPayAddress}>更新支付设置</Form.Button>
          <Divider />
          <Header as='h3' inverted={isDark}>
            配置登录注册
          </Header>
          <Form.Group inline>
            <Form.Checkbox
              checked={inputs.PasswordLoginEnabled === 'true'}
              label='允许通过密码进行登录'
              name='PasswordLoginEnabled'
              onChange={handleInputChange}
            />
            {showPasswordWarningModal && (
              <Modal
                open={showPasswordWarningModal}
                onClose={() => setShowPasswordWarningModal(false)}
                size={'tiny'}
                style={{ maxWidth: '450px' }}
              >
                <Modal.Header>警告</Modal.Header>
                <Modal.Content>
                  <p>
                    取消密码登录将导致所有未绑定其他登录方式的用户（包括管理员）无法通过密码登录，确认取消？
                  </p>
                </Modal.Content>
                <Modal.Actions>
                  <Button onClick={() => setShowPasswordWarningModal(false)}>
                    取消
                  </Button>
                  <Button
                    color='yellow'
                    onClick={async () => {
                      setShowPasswordWarningModal(false);
                      await updateOption('PasswordLoginEnabled', 'false');
                    }}
                  >
                    确定
                  </Button>
                </Modal.Actions>
              </Modal>
            )}
            <Form.Checkbox
              checked={inputs.PasswordRegisterEnabled === 'true'}
              label='允许通过密码进行注册'
              name='PasswordRegisterEnabled'
              onChange={handleInputChange}
            />
            <Form.Checkbox
              checked={inputs.EmailVerificationEnabled === 'true'}
              label='通过密码注册时需要进行邮箱验证'
              name='EmailVerificationEnabled'
              onChange={handleInputChange}
            />
            <Form.Checkbox
              checked={inputs.GitHubOAuthEnabled === 'true'}
              label='允许通过 GitHub 账户登录 & 注册'
              name='GitHubOAuthEnabled'
              onChange={handleInputChange}
            />
            <Form.Checkbox
              checked={inputs.LinuxDOOAuthEnabled === 'true'}
              label='允许通过 LinuxDO 账户登录 & 注册'
              name='LinuxDOOAuthEnabled'
              onChange={handleInputChange}
            />
            <Form.Checkbox
              checked={inputs.WeChatAuthEnabled === 'true'}
              label='允许通过微信登录 & 注册'
              name='WeChatAuthEnabled'
              onChange={handleInputChange}
            />
            <Form.Checkbox
              checked={inputs.TelegramOAuthEnabled === 'true'}
              label='允许通过 Telegram 进行登录'
              name='TelegramOAuthEnabled'
              onChange={handleInputChange}
            />
          </Form.Group>
          <Form.Group inline>
            <Form.Checkbox
              checked={inputs.RegisterEnabled === 'true'}
              label='允许新用户注册（此项为否时，新用户将无法以任何方式进行注册）'
              name='RegisterEnabled'
              onChange={handleInputChange}
            />
            <Form.Checkbox
              checked={inputs.TurnstileCheckEnabled === 'true'}
              label='启用 Turnstile 用户校验'
              name='TurnstileCheckEnabled'
              onChange={handleInputChange}
            />
          </Form.Group>
          <Divider />
          <Header as='h3' inverted={isDark}>
            配置邮箱域名白名单
            <Header.Subheader>
              用以防止恶意用户利用临时邮箱批量注册
            </Header.Subheader>
          </Header>
          <Form.Group widths={3}>
            <Form.Checkbox
              label='启用邮箱域名白名单'
              name='EmailDomainRestrictionEnabled'
              onChange={handleInputChange}
              checked={inputs.EmailDomainRestrictionEnabled === 'true'}
            />
          </Form.Group>
          <Form.Group widths={3}>
            <Form.Checkbox
              label='启用邮箱别名限制（例如：ab.cd@gmail.com）'
              name='EmailAliasRestrictionEnabled'
              onChange={handleInputChange}
              checked={inputs.EmailAliasRestrictionEnabled === 'true'}
            />
          </Form.Group>
          <Form.Group widths={2}>
            <Form.Dropdown
              label='允许的邮箱域名'
              placeholder='允许的邮箱域名'
              name='EmailDomainWhitelist'
              required
              fluid
              multiple
              selection
              onChange={handleInputChange}
              value={inputs.EmailDomainWhitelist}
              autoComplete='new-password'
              options={EmailDomainWhitelist}
            />
            <Form.Input
              label='添加新的允许的邮箱域名'
              action={
                <Button
                  type='button'
                  onClick={() => {
                    submitNewRestrictedDomain();
                  }}
                >
                  填入
                </Button>
              }
              onKeyDown={(e) => {
                if (e.key === 'Enter') {
                  submitNewRestrictedDomain();
                }
              }}
              autoComplete='new-password'
              placeholder='输入新的允许的邮箱域名'
              value={restrictedDomainInput}
              onChange={(e, { value }) => {
                setRestrictedDomainInput(value);
              }}
            />
          </Form.Group>
          <Form.Button onClick={submitEmailDomainWhitelist}>
            保存邮箱域名白名单设置
          </Form.Button>
          <Divider />
          <Header as='h3' inverted={isDark}>
            配置 SMTP
            <Header.Subheader>用以支持系统的邮件发送</Header.Subheader>
          </Header>
          <Form.Group widths={3}>
            <Form.Input
              label='SMTP 服务器地址'
              name='SMTPServer'
              onChange={handleInputChange}
              autoComplete='new-password'
              value={inputs.SMTPServer}
              placeholder='例如：smtp.qq.com'
            />
            <Form.Input
              label='SMTP 端口'
              name='SMTPPort'
              onChange={handleInputChange}
              autoComplete='new-password'
              value={inputs.SMTPPort}
              placeholder='默认: 587'
            />
            <Form.Input
              label='SMTP 账户'
              name='SMTPAccount'
              onChange={handleInputChange}
              autoComplete='new-password'
              value={inputs.SMTPAccount}
              placeholder='通常是邮箱地址'
            />
          </Form.Group>
          <Form.Group widths={3}>
            <Form.Input
              label='SMTP 发送者邮箱'
              name='SMTPFrom'
              onChange={handleInputChange}
              autoComplete='new-password'
              value={inputs.SMTPFrom}
              placeholder='通常和邮箱地址保持一致'
            />
            <Form.Input
              label='SMTP 访问凭证'
              name='SMTPToken'
              onChange={handleInputChange}
              type='password'
              autoComplete='new-password'
              checked={inputs.RegisterEnabled === 'true'}
              placeholder='敏感信息不会发送到前端显示'
            />
          </Form.Group>
          <Form.Group widths={3}>
            <Form.Checkbox
              label='启用SMTP SSL（465端口强制开启）'
              name='SMTPSSLEnabled'
              onChange={handleInputChange}
              checked={inputs.SMTPSSLEnabled === 'true'}
            />
          </Form.Group>
          <Form.Button onClick={submitSMTP}>保存 SMTP 设置</Form.Button>
          <Divider />
          <Header as='h3' inverted={isDark}>
            配置 GitHub OAuth App
            <Header.Subheader>
              用以支持通过 GitHub 进行登录注册，
              <a
                href='https://github.com/settings/developers'
                target='_blank'
                rel='noreferrer'
              >
                点击此处
              </a>
              管理你的 GitHub OAuth App
            </Header.Subheader>
          </Header>
          <Message>
            Homepage URL 填 <code>{inputs.ServerAddress}</code>
            ，Authorization callback URL 填{' '}
            <code>{`${inputs.ServerAddress}/oauth/github`}</code>
          </Message>
          <Form.Group widths={3}>
            <Form.Input
              label='GitHub Client ID'
              name='GitHubClientId'
              onChange={handleInputChange}
              autoComplete='new-password'
              value={inputs.GitHubClientId}
              placeholder='输入你注册的 GitHub OAuth APP 的 ID'
            />
            <Form.Input
              label='GitHub Client Secret'
              name='GitHubClientSecret'
              onChange={handleInputChange}
              type='password'
              autoComplete='new-password'
              value={inputs.GitHubClientSecret}
              placeholder='敏感信息不会发送到前端显示'
            />
          </Form.Group>
          <Form.Button onClick={submitGitHubOAuth}>
            保存 GitHub OAuth 设置
          </Form.Button>
          <Divider />
          <Header as='h3' inverted={isDark}>
            配置 WeChat Server
            <Header.Subheader>
              用以支持通过微信进行登录注册，
              <a
                href='https://github.com/songquanpeng/wechat-server'
                target='_blank'
                rel='noreferrer'
              >
                点击此处
              </a>
              了解 WeChat Server
            </Header.Subheader>
          </Header>
          <Form.Group widths={3}>
            <Form.Input
              label='WeChat Server 服务器地址'
              name='WeChatServerAddress'
              placeholder='例如：https://yourdomain.com'
              onChange={handleInputChange}
              autoComplete='new-password'
              value={inputs.WeChatServerAddress}
            />
            <Form.Input
              label='WeChat Server 访问凭证'
              name='WeChatServerToken'
              type='password'
              onChange={handleInputChange}
              autoComplete='new-password'
              value={inputs.WeChatServerToken}
              placeholder='敏感信息不会发送到前端显示'
            />
            <Form.Input
              label='微信公众号二维码图片链接'
              name='WeChatAccountQRCodeImageURL'
              onChange={handleInputChange}
              autoComplete='new-password'
              value={inputs.WeChatAccountQRCodeImageURL}
              placeholder='输入一个图片链接'
            />
          </Form.Group>
          <Form.Button onClick={submitWeChat}>
            保存 WeChat Server 设置
          </Form.Button>
          <Divider />
          <Header as='h3' inverted={isDark}>
            配置 Telegram 登录
          </Header>
          <Form.Group inline>
            <Form.Input
              label='Telegram Bot Token'
              name='TelegramBotToken'
              onChange={handleInputChange}
              value={inputs.TelegramBotToken}
              placeholder='输入你的 Telegram Bot Token'
            />
            <Form.Input
              label='Telegram Bot 名称'
              name='TelegramBotName'
              onChange={handleInputChange}
              value={inputs.TelegramBotName}
              placeholder='输入你的 Telegram Bot 名称'
            />
          </Form.Group>
          <Form.Button onClick={submitTelegramSettings}>
            保存 Telegram 登录设置
          </Form.Button>
          <Divider />
          <Header as='h3' inverted={isDark}>
            配置 Turnstile
            <Header.Subheader>
              用以支持用户校验，
              <a
                href='https://dash.cloudflare.com/'
                target='_blank'
                rel='noreferrer'
              >
                点击此处
              </a>
              管理你的 Turnstile Sites，推荐选择 Invisible Widget Type
            </Header.Subheader>
          </Header>
          <Form.Group widths={3}>
            <Form.Input
              label='Turnstile Site Key'
              name='TurnstileSiteKey'
              onChange={handleInputChange}
              autoComplete='new-password'
              value={inputs.TurnstileSiteKey}
              placeholder='输入你注册的 Turnstile Site Key'
            />
            <Form.Input
              label='Turnstile Secret Key'
              name='TurnstileSecretKey'
              onChange={handleInputChange}
              type='password'
              autoComplete='new-password'
              value={inputs.TurnstileSecretKey}
              placeholder='敏感信息不会发送到前端显示'
            />
          </Form.Group>
          <Form.Button onClick={submitTurnstile}>
            保存 Turnstile 设置
          </Form.Button>
          <Divider />
          <Header as='h3' inverted={isDark}>
            配置 LinuxDO OAuth App
            <Header.Subheader>
              用以支持通过 LinuxDO 进行登录注册，
              <a
                href='https://connect.linux.do/'
                target='_blank'
                rel='noreferrer'
              >
                点击此处
              </a>
              管理你的 LinuxDO OAuth App
            </Header.Subheader>
          </Header>
          <Message>
            Homepage URL 填 <code>{inputs.ServerAddress}</code>
            ，Authorization callback URL 填{' '}
            <code>{`${inputs.ServerAddress}/oauth/linuxdo`}</code>
          </Message>
          <Form.Group widths={3}>
            <Form.Input
              label='LinuxDO Client ID'
              name='LinuxDOClientId'
              onChange={handleInputChange}
              autoComplete='new-password'
              value={inputs.LinuxDOClientId}
              placeholder='输入你注册的 LinuxDO OAuth APP 的 ID'
            />
            <Form.Input
              label='LinuxDO Client Secret'
              name='LinuxDOClientSecret'
              onChange={handleInputChange}
              type='password'
              autoComplete='new-password'
              value={inputs.LinuxDOClientSecret}
              placeholder='敏感信息不会发送到前端显示'
            />
          </Form.Group>
          <Form.Button onClick={submitLinuxDOOAuth}>
            保存 LinuxDO OAuth 设置
          </Form.Button>
        </Form>
      </Grid.Column>
    </Grid>
  );
};

export default SystemSetting;
