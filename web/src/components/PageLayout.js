import HeaderBar from './HeaderBar.js';
import { Layout } from '@douyinfe/semi-ui';
import SiderBar from './SiderBar.js';
import App from '../App.js';
import FooterBar from './Footer.js';
import { ToastContainer } from 'react-toastify';
import React, { useContext } from 'react';
import { StyleContext } from '../context/Style/index.js';
const { Sider, Content, Header, Footer } = Layout;


const PageLayout = () => {
  const [styleState, styleDispatch] = useContext(StyleContext);

  return (
    <Layout style={{ height: '100vh', display: 'flex', flexDirection: 'column' }}>
      <Header>
        <HeaderBar />
      </Header>
      <Layout style={{ flex: 1, overflow: 'hidden' }}>
        <Sider>
          {styleState.showSider ? null : <SiderBar />}
        </Sider>
        <Layout>
          <Content
            style={{ overflowY: styleState.shouldInnerPadding?'hidden':'auto', padding: styleState.shouldInnerPadding? '0': '24px' }}
          >
            <App />
          </Content>
          <Layout.Footer>
            <FooterBar></FooterBar>
          </Layout.Footer>
        </Layout>
      </Layout>
      <ToastContainer />
    </Layout>
  )
}

export default PageLayout;