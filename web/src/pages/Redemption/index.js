import React from 'react';
import { Segment, Header } from 'semantic-ui-react';
import RedemptionsTable from '../../components/RedemptionsTable';
import TokensTable from "../../components/TokensTable";
import {Layout} from "@douyinfe/semi-ui";

const Redemption = () => (
  <>
      <Layout>
          <Layout.Header>
              <h3>管理兑换码</h3>
          </Layout.Header>
          <Layout.Content>
              <RedemptionsTable/>
          </Layout.Content>
      </Layout>
  </>
);

export default Redemption;
