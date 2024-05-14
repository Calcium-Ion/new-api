import React, { useEffect, useState } from 'react';
import { Card, Spin } from '@douyinfe/semi-ui';
import SettingsGeneral from '../pages/Setting/Operation/SettingsGeneral.js';
import SettingsDrawing from '../pages/Setting/Operation/SettingsDrawing.js';
import SettingsSensitiveWords from '../pages/Setting/Operation/SettingsSensitiveWords.js';
import SettingsLog from '../pages/Setting/Operation/SettingsLog.js';
import SettingsDataDashboard from '../pages/Setting/Operation/SettingsDataDashboard.js';
import SettingsMonitoring from '../pages/Setting/Operation/SettingsMonitoring.js';
import SettingsCreditLimit from '../pages/Setting/Operation/SettingsCreditLimit.js';
import SettingsMagnification from '../pages/Setting/Operation/SettingsMagnification.js';

import { API, showError, showSuccess } from '../helpers';

const OperationSetting = () => {
  let [inputs, setInputs] = useState({
    QuotaForNewUser: 0,
    QuotaForInviter: 0,
    QuotaForInvitee: 0,
    QuotaRemindThreshold: 0,
    PreConsumedQuota: 0,
    StreamCacheQueueLength: 0,
    ModelRatio: '',
    CompletionRatio: '',
    ModelPrice: '',
    GroupRatio: '',
    TopUpLink: '',
    ChatLink: '',
    ChatLink2: '', // 添加的新状态变量
    QuotaPerUnit: 0,
    AutomaticDisableChannelEnabled: false,
    AutomaticEnableChannelEnabled: false,
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
          item.key === 'CompletionRatio' ||
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
    } else {
      showError(message);
    }
  };
  async function onRefresh() {
    try {
      setLoading(true);
      await getOptions();
      showSuccess('刷新成功');
    } catch (error) {
      showError('刷新失败');
    } finally {
      setLoading(false);
    }
  }

  useEffect(() => {
    onRefresh();
  }, []);

  return (
    <>
      <Spin spinning={loading} size='large'>
        {/* 通用设置 */}
        <Card style={{ marginTop: '10px' }}>
          <SettingsGeneral options={inputs} refresh={onRefresh} />
        </Card>
        {/* 绘图设置 */}
        <Card style={{ marginTop: '10px' }}>
          <SettingsDrawing options={inputs} refresh={onRefresh} />
        </Card>
        {/* 屏蔽词过滤设置 */}
        <Card style={{ marginTop: '10px' }}>
          <SettingsSensitiveWords options={inputs} refresh={onRefresh} />
        </Card>
        {/* 日志设置 */}
        <Card style={{ marginTop: '10px' }}>
          <SettingsLog options={inputs} refresh={onRefresh} />
        </Card>
        {/* 数据看板 */}
        <Card style={{ marginTop: '10px' }}>
          <SettingsDataDashboard options={inputs} refresh={onRefresh} />
        </Card>
        {/* 监控设置 */}
        <Card style={{ marginTop: '10px' }}>
          <SettingsMonitoring options={inputs} refresh={onRefresh} />
        </Card>
        {/* 额度设置 */}
        <Card style={{ marginTop: '10px' }}>
          <SettingsCreditLimit options={inputs} refresh={onRefresh} />
        </Card>
        {/* 倍率设置 */}
        <Card style={{ marginTop: '10px' }}>
          <SettingsMagnification options={inputs} refresh={onRefresh} />
        </Card>
      </Spin>
    </>
  );
};

export default OperationSetting;
