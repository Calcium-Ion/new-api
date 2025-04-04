import React from 'react';
import RedemptionsTable from '../../components/RedemptionsTable';
import { Layout } from '@douyinfe/semi-ui';
import { useTranslation } from 'react-i18next';

const Redemption = () => {
  const { t } = useTranslation();
  return (
    <>
      <Layout>
        <Layout.Header>
          <h3>{t('管理兑换码')}</h3>
        </Layout.Header>
        <Layout.Content>
          <RedemptionsTable />
        </Layout.Content>
      </Layout>
    </>
  );
};

export default Redemption;
