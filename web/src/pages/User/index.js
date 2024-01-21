import React from 'react';
import UsersTable from '../../components/UsersTable';
import {Layout} from "@douyinfe/semi-ui";

const User = () => (
  <>
    <Layout>
        <Layout.Header>
            <h3>管理用户</h3>
        </Layout.Header>
        <Layout.Content>
            <UsersTable/>
        </Layout.Content>
    </Layout>
  </>
);

export default User;
