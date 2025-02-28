import React, { useEffect, useState } from 'react';
import { Card, Spin, Tabs } from '@douyinfe/semi-ui';


import { API, showError, showSuccess } from '../helpers';
import { useTranslation } from 'react-i18next';
import SettingGeminiModel from '../pages/Setting/Model/SettingGeminiModel.js';
import SettingClaudeModel from '../pages/Setting/Model/SettingClaudeModel.js';

const ModelSetting = () => {
  const { t } = useTranslation();
  let [inputs, setInputs] = useState({
    'gemini.safety_settings': '',
    'gemini.version_settings': '',
    'claude.model_headers_settings': '',
    'claude.thinking_adapter_enabled': true,
    'claude.default_max_tokens': '',
    'claude.thinking_adapter_budget_tokens_percentage': 0.8,
  });

  let [loading, setLoading] = useState(false);

  const getOptions = async () => {
    const res = await API.get('/api/option/');
    const { success, message, data } = res.data;
    if (success) {
      let newInputs = {};
      data.forEach((item) => {
        if (
          item.key === 'gemini.safety_settings' ||
          item.key === 'gemini.version_settings' ||
          item.key === 'claude.model_headers_settings'||
          item.key === 'claude.default_max_tokens'
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
        {/* Claude */}
        <Card style={{ marginTop: '10px' }}>
          <SettingClaudeModel options={inputs} refresh={onRefresh} />
        </Card>
      </Spin>
    </>
  );
};

export default ModelSetting;
