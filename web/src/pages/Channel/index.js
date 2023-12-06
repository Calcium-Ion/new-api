import React from 'react';
import ChannelsTable from '../../components/ChannelsTable';
import {Layout} from "@douyinfe/semi-ui";
import RedemptionsTable from "../../components/RedemptionsTable";

const File = () => (
    <>
        <Layout>
            <Layout.Header>
                <h3>管理渠道</h3>
            </Layout.Header>
            <Layout.Content>
                <ChannelsTable/>
            </Layout.Content>
        </Layout>
    </>
);

export default File;
