import React from 'react';
import TokensTable from '../../components/TokensTable';
import {Layout} from "@douyinfe/semi-ui";
const Token = () => (
  <>
    <Layout>
      <Layout.Header>
          <h3>我的令牌</h3>
      </Layout.Header>
      <Layout.Content>
          <TokensTable/>
      </Layout.Content>
    </Layout>
  </>
);

export default Token;
