import HeaderBar from './HeaderBar.js';
import { Layout } from '@douyinfe/semi-ui';
import SiderBar from './SiderBar.js';
import App from '../App.js';
import FooterBar from './Footer.js';
import { ToastContainer } from 'react-toastify';
import React, { useContext } from 'react';
import { StyleContext } from '../context/Style/index.js';
import { useTranslation } from 'react-i18next';
const { Sider, Content, Header, Footer } = Layout;


const PageLayout = () => {
  const [styleState, styleDispatch] = useContext(StyleContext);
  const { t } = useTranslation();

  return (
    <Layout style={{ height: '100vh', display: 'flex', flexDirection: 'column' }}>
      <Header>
        <HeaderBar />
      </Header>
      <Layout style={{ flex: 1, overflow: 'hidden' }}>
        <Sider>
          {styleState.showSider ? <SiderBar /> : null}
        </Sider>
        <Layout>
          <Content
            style={{ overflowY: 'auto', padding: styleState.shouldInnerPadding? '24px': '0' }}
          >
            <App />
          </Content>
          <Layout.Footer>
            <FooterBar />
          </Layout.Footer>
        </Layout>
      </Layout>
      <ToastContainer />
    </Layout>
  )
}

export default PageLayout;