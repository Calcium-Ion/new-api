import React, { useEffect, useState } from 'react';
import { Card, Spin, Tabs } from '@douyinfe/semi-ui';
import SettingsGeneral from '../pages/Setting/Operation/SettingsGeneral.js';
import SettingsDrawing from '../pages/Setting/Operation/SettingsDrawing.js';
import SettingsSensitiveWords from '../pages/Setting/Operation/SettingsSensitiveWords.js';
import SettingsLog from '../pages/Setting/Operation/SettingsLog.js';
import SettingsDataDashboard from '../pages/Setting/Operation/SettingsDataDashboard.js';
import SettingsMonitoring from '../pages/Setting/Operation/SettingsMonitoring.js';
import SettingsCreditLimit from '../pages/Setting/Operation/SettingsCreditLimit.js';
import SettingsMagnification from '../pages/Setting/Operation/SettingsMagnification.js';
import ModelSettingsVisualEditor from '../pages/Setting/Operation/ModelSettingsVisualEditor.js';
import GroupRatioSettings from '../pages/Setting/Operation/GroupRatioSettings.js';
import ModelRatioSettings from '../pages/Setting/Operation/ModelRatioSettings.js';


import { API, showError, showSuccess } from '../helpers';
import SettingsChats from '../pages/Setting/Operation/SettingsChats.js';
import { useTranslation } from 'react-i18next';
import RequestRateLimit from '../pages/Setting/RateLimit/SettingsRequestRateLimit.js';

const RateLimitSetting = () => {
  const { t } = useTranslation();
  let [inputs, setInputs] = useState({
    ModelRequestRateLimitEnabled: false,
    ModelRequestRateLimitCount: 0,
    ModelRequestRateLimitSuccessCount: 1000,
    ModelRequestRateLimitDurationMinutes: 1,
  });

  let [loading, setLoading] = useState(false);

  const getOptions = async () => {
    const res = await API.get('/api/option/');
    const { success, message, data } = res.data;
    if (success) {
      let newInputs = {};
      data.forEach((item) => {
        if (
          item.key.endsWith('Enabled')
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
      // showSuccess('刷新成功');
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
        {/* AI请求速率限制 */}
        <Card style={{ marginTop: '10px' }}>
          <RequestRateLimit options={inputs} refresh={onRefresh} />
        </Card>
      </Spin>
    </>
  );
};

export default RateLimitSetting;
