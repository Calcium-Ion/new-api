import React, { useEffect, useState } from 'react';
import { Layout, TabPane, Tabs } from '@douyinfe/semi-ui';
import { useNavigate, useLocation } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

import SystemSetting from '../../components/SystemSetting';
import { isRoot } from '../../helpers';
import OtherSetting from '../../components/OtherSetting';
import PersonalSetting from '../../components/PersonalSetting';
import OperationSetting from '../../components/OperationSetting';
import RateLimitSetting from '../../components/RateLimitSetting.js';
import ModelSetting from '../../components/ModelSetting.js';

const Setting = () => {
  const { t } = useTranslation();
  const navigate = useNavigate();
  const location = useLocation();
  const [tabActiveKey, setTabActiveKey] = useState('1');
  let panes = [];

  if (isRoot()) {
    panes.push({
      tab: t('运营设置'),
      content: <OperationSetting />,
      itemKey: 'operation',
    });
    panes.push({
      tab: t('速率限制设置'),
      content: <RateLimitSetting />,
      itemKey: 'ratelimit',
    });
    panes.push({
      tab: t('模型相关设置'),
      content: <ModelSetting />,
      itemKey: 'models',
    });
    panes.push({
      tab: t('系统设置'),
      content: <SystemSetting />,
      itemKey: 'system',
    });
    panes.push({
      tab: t('其他设置'),
      content: <OtherSetting />,
      itemKey: 'other',
    });
  }
  const onChangeTab = (key) => {
    setTabActiveKey(key);
    navigate(`?tab=${key}`);
  };
  useEffect(() => {
    const searchParams = new URLSearchParams(window.location.search);
    const tab = searchParams.get('tab');
    if (tab) {
      setTabActiveKey(tab);
    } else {
      onChangeTab('operation');
    }
  }, [location.search]);
  return (
    <div>
      <Layout>
        <Layout.Content>
          <Tabs
            type='line'
            activeKey={tabActiveKey}
            onChange={(key) => onChangeTab(key)}
          >
            {panes.map((pane) => (
              <TabPane itemKey={pane.itemKey} tab={pane.tab} key={pane.itemKey}>
                {tabActiveKey === pane.itemKey && pane.content}
              </TabPane>
            ))}
          </Tabs>
        </Layout.Content>
      </Layout>
    </div>
  );
};

export default Setting;
