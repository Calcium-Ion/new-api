import React, { useEffect, useState } from 'react';
import { Card, Spin, Tabs } from '@douyinfe/semi-ui';


import { API, showError, showSuccess } from '../helpers';
import SettingsChats from '../pages/Setting/Operation/SettingsChats.js';
import { useTranslation } from 'react-i18next';
import SettingGeminiModel from '../pages/Setting/Model/SettingGeminiModel.js';

const ModelSetting = () => {
  const { t } = useTranslation();
  let [inputs, setInputs] = useState({
    GeminiSafetySettings: '',
    GeminiVersionSettings: '',
  });

  let [loading, setLoading] = useState(false);

  const getOptions = async () => {
    const res = await API.get('/api/option/');
    const { success, message, data } = res.data;
    if (success) {
      let newInputs = {};
      data.forEach((item) => {
        if (
          item.key === 'GeminiSafetySettings' ||
          item.key === 'GeminiVersionSettings'
        ) {
          item.value = JSON.stringify(JSON.parse(item.value), null, 2);
        }
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
        {/* Gemini */}
        <Card style={{ marginTop: '10px' }}>
          <SettingGeminiModel options={inputs} refresh={onRefresh} />
        </Card>
      </Spin>
    </>
  );
};

export default ModelSetting;
