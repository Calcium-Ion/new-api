import React from 'react';
import TokensTable from '../../components/TokensTable';
import { Banner, Layout } from '@douyinfe/semi-ui';
const Token = () => (
  <>
    <Layout>
      <Layout.Header>
        <Banner
          type='warning'
          description='令牌无法精确计算使用额度，请勿直接将令牌分享他人。'
        />
      </Layout.Header>
      <Layout.Content>
        <TokensTable />
      </Layout.Content>
    </Layout>
  </>
);

export default Token;
