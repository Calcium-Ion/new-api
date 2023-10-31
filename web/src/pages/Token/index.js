import React from 'react';
import TokensTable from '../../components/TokensTable';
import {Layout} from "@douyinfe/semi-ui";
const {Content, Header} = Layout;
const Token = () => (
  <>
    <Layout>
      <Header>
          <h3>我的令牌</h3>
      </Header>
      <Content>
          <TokensTable/>
      </Content>
    </Layout>
  </>
);

export default Token;
