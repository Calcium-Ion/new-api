import HeaderBar from './HeaderBar.js';
import { Layout } from '@douyinfe/semi-ui';
import SiderBar from './SiderBar.js';
import App from '../App.js';
import FooterBar from './Footer.js';
import { ToastContainer } from 'react-toastify';
import React, { useContext, useEffect } from 'react';
import { StyleContext } from '../context/Style/index.js';
import { useTranslation } from 'react-i18next';
import { API, getLogo, getSystemName, showError } from '../helpers/index.js';
import { setStatusData } from '../helpers/data.js';
import { UserContext } from '../context/User/index.js';
import { StatusContext } from '../context/Status/index.js';
const { Sider, Content, Header, Footer } = Layout;


const PageLayout = () => {
  const [userState, userDispatch] = useContext(UserContext);
  const [statusState, statusDispatch] = useContext(StatusContext);
  const [styleState, styleDispatch] = useContext(StyleContext);
  const { i18n } = useTranslation();

  const loadUser = () => {
    let user = localStorage.getItem('user');
    if (user) {
      let data = JSON.parse(user);
      userDispatch({ type: 'login', payload: data });
    }
  };

  const loadStatus = async () => {
    try {
      const res = await API.get('/api/status');
      const { success, data } = res.data;
      if (success) {
        statusDispatch({ type: 'set', payload: data });
        setStatusData(data);
      } else {
        showError('Unable to connect to server');
      }
    } catch (error) {
      showError('Failed to load status');
    }
  };

  useEffect(() => {
    loadUser();
    loadStatus().catch(console.error);
    let systemName = getSystemName();
    if (systemName) {
      document.title = systemName;
    }
    let logo = getLogo();
    if (logo) {
      let linkElement = document.querySelector("link[rel~='icon']");
      if (linkElement) {
        linkElement.href = logo;
      }
    }
    // 从localStorage获取上次使用的语言
    const savedLang = localStorage.getItem('i18nextLng');
    if (savedLang) {
      i18n.changeLanguage(savedLang);
    }
  }, [i18n]);

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