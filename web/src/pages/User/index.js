import React from 'react';
import UsersTable from '../../components/UsersTable';
import { Layout } from '@douyinfe/semi-ui';
import { useTranslation } from 'react-i18next';

const User = () => {
  const { t } = useTranslation();
  return (
    <>
      <Layout>
        <Layout.Header>
        <h3>{t('管理用户')}</h3>
      </Layout.Header>
      <Layout.Content>
        <UsersTable />
        </Layout.Content>
      </Layout>
    </>
  );
};

export default User;
