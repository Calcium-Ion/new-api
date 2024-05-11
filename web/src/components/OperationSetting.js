import React, { useEffect, useState } from 'react';
import { Divider, Form, Grid, Header } from 'semantic-ui-react';
import { Card } from '@douyinfe/semi-ui';
import SettingsGeneral from '../pages/Setting/Operation/SettingsGeneral.js';
import SettingsDrawing from '../pages/Setting/Operation/SettingsDrawing.js';
import SettingsSensitiveWords from '../pages/Setting/Operation/SettingsSensitiveWords.js';
import SettingsLog from '../pages/Setting/Operation/SettingsLog.js';
import SettingsDataDashboard from '../pages/Setting/Operation/SettingsDataDashboard.js';

import {
  API,
  showError,
  showSuccess,
  timestamp2string,
  verifyJSON,
} from '../helpers';

import { useTheme } from '../context/Theme';

const OperationSetting = () => {
  let now = new Date();
  let [inputs, setInputs] = useState({
    QuotaForNewUser: 0,
    QuotaForInviter: 0,
    QuotaForInvitee: 0,
    QuotaRemindThreshold: 0,
    PreConsumedQuota: 0,
    StreamCacheQueueLength: 0,
    ModelRatio: '',
    ModelPrice: '',
    GroupRatio: '',
    TopUpLink: '',
    ChatLink: '',
    ChatLink2: '', // 添加的新状态变量
    QuotaPerUnit: 0,
    AutomaticDisableChannelEnabled: '',
    AutomaticEnableChannelEnabled: '',
    ChannelDisableThreshold: 0,
    LogConsumeEnabled: false,
    DisplayInCurrencyEnabled: false,
    DisplayTokenStatEnabled: false,
    CheckSensitiveEnabled: false,
    CheckSensitiveOnPromptEnabled: false,
    CheckSensitiveOnCompletionEnabled: '',
    StopOnSensitiveEnabled: '',
    SensitiveWords: '',
    MjNotifyEnabled: false,
    MjAccountFilterEnabled: false,
    MjModeClearEnabled: false,
    MjForwardUrlEnabled: false,
    DrawingEnabled: false,
    DataExportEnabled: false,
    DataExportDefaultTime: 'hour',
    DataExportInterval: 5,
    DefaultCollapseSidebar: false, // 默认折叠侧边栏
    RetryTimes: 0,
  });
  const [originInputs, setOriginInputs] = useState({});
  let [loading, setLoading] = useState(false);

  const getOptions = async () => {
    const res = await API.get('/api/option/');
    const { success, message, data } = res.data;
    if (success) {
      let newInputs = {};
      data.forEach((item) => {
        if (
          item.key === 'ModelRatio' ||
          item.key === 'GroupRatio' ||
          item.key === 'ModelPrice'
        ) {
          item.value = JSON.stringify(JSON.parse(item.value), null, 2);
        }
        if (
          item.key.endsWith('Enabled') ||
          ['DefaultCollapseSidebar'].includes(item.key)
        ) {
          newInputs[item.key] = item.value === 'true' ? true : false;
        } else {
          newInputs[item.key] = item.value;
        }
      });

      setInputs(newInputs);
      setOriginInputs(newInputs);
    } else {
      showError(message);
    }
  };

  const theme = useTheme();
  const isDark = theme === 'dark';

  useEffect(() => {
    getOptions().then();
  }, []);

  const updateOption = async (key, value) => {
    setLoading(true);
    if (key.endsWith('Enabled')) {
      value = inputs[key] === 'true' ? 'false' : 'true';
    }
    if (key === 'DefaultCollapseSidebar') {
      value = inputs[key] === 'true' ? 'false' : 'true';
    }
    console.log(key, value);
    const res = await API.put('/api/option/', {
      key,
      value,
    });
    const { success, message } = res.data;
    if (success) {
      setInputs((inputs) => ({ ...inputs, [key]: value }));
    } else {
      showError(message);
    }
    setLoading(false);
  };

  const handleInputChange = async (e, { name, value }) => {
    if (
      name.endsWith('Enabled') ||
      name === 'DataExportInterval' ||
      name === 'DataExportDefaultTime' ||
      name === 'DefaultCollapseSidebar'
    ) {
      if (name === 'DataExportDefaultTime') {
        localStorage.setItem('data_export_default_time', value);
      } else if (name === 'MjNotifyEnabled') {
        localStorage.setItem('mj_notify_enabled', value);
      }
      await updateOption(name, value);
    } else {
      setInputs((inputs) => ({ ...inputs, [name]: value }));
    }
  };

  const submitConfig = async (group) => {
    switch (group) {
      case 'monitor':
        if (
          originInputs['ChannelDisableThreshold'] !==
          inputs.ChannelDisableThreshold
        ) {
          await updateOption(
            'ChannelDisableThreshold',
            inputs.ChannelDisableThreshold,
          );
        }
        if (
          originInputs['QuotaRemindThreshold'] !== inputs.QuotaRemindThreshold
        ) {
          await updateOption(
            'QuotaRemindThreshold',
            inputs.QuotaRemindThreshold,
          );
        }
        break;
      case 'ratio':
        if (originInputs['ModelRatio'] !== inputs.ModelRatio) {
          if (!verifyJSON(inputs.ModelRatio)) {
            showError('模型倍率不是合法的 JSON 字符串');
            return;
          }
          await updateOption('ModelRatio', inputs.ModelRatio);
        }
        if (originInputs['GroupRatio'] !== inputs.GroupRatio) {
          if (!verifyJSON(inputs.GroupRatio)) {
            showError('分组倍率不是合法的 JSON 字符串');
            return;
          }
          await updateOption('GroupRatio', inputs.GroupRatio);
        }
        if (originInputs['ModelPrice'] !== inputs.ModelPrice) {
          if (!verifyJSON(inputs.ModelPrice)) {
            showError('模型固定价格不是合法的 JSON 字符串');
            return;
          }
          await updateOption('ModelPrice', inputs.ModelPrice);
        }
        break;
      case 'words':
        if (originInputs['SensitiveWords'] !== inputs.SensitiveWords) {
          await updateOption('SensitiveWords', inputs.SensitiveWords);
        }
        break;
      case 'quota':
        if (originInputs['QuotaForNewUser'] !== inputs.QuotaForNewUser) {
          await updateOption('QuotaForNewUser', inputs.QuotaForNewUser);
        }
        if (originInputs['QuotaForInvitee'] !== inputs.QuotaForInvitee) {
          await updateOption('QuotaForInvitee', inputs.QuotaForInvitee);
        }
        if (originInputs['QuotaForInviter'] !== inputs.QuotaForInviter) {
          await updateOption('QuotaForInviter', inputs.QuotaForInviter);
        }
        if (originInputs['PreConsumedQuota'] !== inputs.PreConsumedQuota) {
          await updateOption('PreConsumedQuota', inputs.PreConsumedQuota);
        }
        break;
      case 'general':
        if (originInputs['TopUpLink'] !== inputs.TopUpLink) {
          await updateOption('TopUpLink', inputs.TopUpLink);
        }
        if (originInputs['ChatLink'] !== inputs.ChatLink) {
          await updateOption('ChatLink', inputs.ChatLink);
        }
        if (originInputs['ChatLink2'] !== inputs.ChatLink2) {
          await updateOption('ChatLink2', inputs.ChatLink2);
        }
        if (originInputs['QuotaPerUnit'] !== inputs.QuotaPerUnit) {
          await updateOption('QuotaPerUnit', inputs.QuotaPerUnit);
        }
        if (originInputs['RetryTimes'] !== inputs.RetryTimes) {
          await updateOption('RetryTimes', inputs.RetryTimes);
        }
        break;
    }
  };
  return (
    <>
      {/* 通用设置 */}
      <Card>
        <SettingsGeneral options={inputs} />
      </Card>
      {/* 绘图设置 */}
      <Card style={{ marginTop: '10px' }}>
        <SettingsDrawing options={inputs} />
      </Card>
      {/* 屏蔽词过滤设置 */}
      <Card style={{ marginTop: '10px' }}>
        <SettingsSensitiveWords options={inputs} />
      </Card>
      {/* 日志设置 */}
      <Card style={{ marginTop: '10px' }}>
        <SettingsLog options={inputs} />
      </Card>
      {/* 数据看板 */}
      <Card style={{ marginTop: '10px' }}>
        <SettingsDataDashboard options={inputs} />
      </Card>
      <Grid columns={1}>
        <Grid.Column>
          <Form loading={loading} inverted={isDark}>
            {/*<Form.Group inline>*/}
            {/*  <Form.Checkbox*/}
            {/*    checked={inputs.StopOnSensitiveEnabled === 'true'}*/}
            {/*    label='在检测到屏蔽词时，立刻停止生成，否则替换屏蔽词'*/}
            {/*    name='StopOnSensitiveEnabled'*/}
            {/*    onChange={handleInputChange}*/}
            {/*  />*/}
            {/*</Form.Group>*/}
            {/*<Form.Group>*/}
            {/*  <Form.Input*/}
            {/*    label="流模式下缓存队列，默认不缓存，设置越大检测越准确，但是回复会有卡顿感"*/}
            {/*    name="StreamCacheTextLength"*/}
            {/*    onChange={handleInputChange}*/}
            {/*    value={inputs.StreamCacheQueueLength}*/}
            {/*    type="number"*/}
            {/*    min="0"*/}
            {/*    placeholder="例如：10"*/}
            {/*  />*/}
            {/*</Form.Group>*/}

            <Divider />
            <Header as='h3' inverted={isDark}>
              监控设置
            </Header>
            <Form.Group widths={3}>
              <Form.Input
                label='最长响应时间'
                name='ChannelDisableThreshold'
                onChange={handleInputChange}
                autoComplete='new-password'
                value={inputs.ChannelDisableThreshold}
                type='number'
                min='0'
                placeholder='单位秒，当运行通道全部测试时，超过此时间将自动禁用通道'
              />
              <Form.Input
                label='额度提醒阈值'
                name='QuotaRemindThreshold'
                onChange={handleInputChange}
                autoComplete='new-password'
                value={inputs.QuotaRemindThreshold}
                type='number'
                min='0'
                placeholder='低于此额度时将发送邮件提醒用户'
              />
            </Form.Group>
            <Form.Group inline>
              <Form.Checkbox
                checked={inputs.AutomaticDisableChannelEnabled === 'true'}
                label='失败时自动禁用通道'
                name='AutomaticDisableChannelEnabled'
                onChange={handleInputChange}
              />
              <Form.Checkbox
                checked={inputs.AutomaticEnableChannelEnabled === 'true'}
                label='成功时自动启用通道'
                name='AutomaticEnableChannelEnabled'
                onChange={handleInputChange}
              />
            </Form.Group>
            <Form.Button
              onClick={() => {
                submitConfig('monitor').then();
              }}
            >
              保存监控设置
            </Form.Button>
            <Divider />
            <Header as='h3' inverted={isDark}>
              额度设置
            </Header>
            <Form.Group widths={4}>
              <Form.Input
                label='新用户初始额度'
                name='QuotaForNewUser'
                onChange={handleInputChange}
                autoComplete='new-password'
                value={inputs.QuotaForNewUser}
                type='number'
                min='0'
                placeholder='例如：100'
              />
              <Form.Input
                label='请求预扣费额度'
                name='PreConsumedQuota'
                onChange={handleInputChange}
                autoComplete='new-password'
                value={inputs.PreConsumedQuota}
                type='number'
                min='0'
                placeholder='请求结束后多退少补'
              />
              <Form.Input
                label='邀请新用户奖励额度'
                name='QuotaForInviter'
                onChange={handleInputChange}
                autoComplete='new-password'
                value={inputs.QuotaForInviter}
                type='number'
                min='0'
                placeholder='例如：2000'
              />
              <Form.Input
                label='新用户使用邀请码奖励额度'
                name='QuotaForInvitee'
                onChange={handleInputChange}
                autoComplete='new-password'
                value={inputs.QuotaForInvitee}
                type='number'
                min='0'
                placeholder='例如：1000'
              />
            </Form.Group>
            <Form.Button
              onClick={() => {
                submitConfig('quota').then();
              }}
            >
              保存额度设置
            </Form.Button>
            <Divider />
            <Header as='h3' inverted={isDark}>
              倍率设置
            </Header>
            <Form.Group widths='equal'>
              <Form.TextArea
                label='模型固定价格（一次调用消耗多少刀，优先级大于模型倍率）'
                name='ModelPrice'
                onChange={handleInputChange}
                style={{
                  minHeight: 250,
                  fontFamily: 'JetBrains Mono, Consolas',
                }}
                autoComplete='new-password'
                value={inputs.ModelPrice}
                placeholder='为一个 JSON 文本，键为模型名称，值为一次调用消耗多少刀，比如 "gpt-4-gizmo-*": 0.1，一次消耗0.1刀'
              />
            </Form.Group>
            <Form.Group widths='equal'>
              <Form.TextArea
                label='模型倍率'
                name='ModelRatio'
                onChange={handleInputChange}
                style={{
                  minHeight: 250,
                  fontFamily: 'JetBrains Mono, Consolas',
                }}
                autoComplete='new-password'
                value={inputs.ModelRatio}
                placeholder='为一个 JSON 文本，键为模型名称，值为倍率'
              />
            </Form.Group>
            <Form.Group widths='equal'>
              <Form.TextArea
                label='分组倍率'
                name='GroupRatio'
                onChange={handleInputChange}
                style={{
                  minHeight: 250,
                  fontFamily: 'JetBrains Mono, Consolas',
                }}
                autoComplete='new-password'
                value={inputs.GroupRatio}
                placeholder='为一个 JSON 文本，键为分组名称，值为倍率'
              />
            </Form.Group>
            <Form.Button
              onClick={() => {
                submitConfig('ratio').then();
              }}
            >
              保存倍率设置
            </Form.Button>
          </Form>
        </Grid.Column>
      </Grid>
    </>
  );
};

export default OperationSetting;
