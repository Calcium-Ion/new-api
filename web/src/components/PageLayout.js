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

    // 默认显示侧边栏
    styleDispatch({ type: 'SET_SIDER', payload: true });
  }, [i18n]);

  // 获取侧边栏折叠状态
  const isSidebarCollapsed =
    localStorage.getItem('default_collapse_sidebar') === 'true';

  return (
    <Layout
      style={{
        height: '100vh',
        display: 'flex',
        flexDirection: 'column',
        overflow: styleState.isMobile ? 'visible' : 'hidden',
      }}
    >
      <Header
        style={{
          padding: 0,
          height: 'auto',
          lineHeight: 'normal',
          position: styleState.isMobile ? 'sticky' : 'fixed',
          width: '100%',
          top: 0,
          zIndex: 100,
          boxShadow: '0 1px 6px rgba(0, 0, 0, 0.08)',
        }}
      >
        <HeaderBar />
      </Header>
      <Layout
        style={{
          marginTop: styleState.isMobile ? '0' : '56px',
          height: styleState.isMobile ? 'auto' : 'calc(100vh - 56px)',
          overflow: styleState.isMobile ? 'visible' : 'auto',
          display: 'flex',
          flexDirection: 'column',
        }}
      >
        {styleState.showSider && (
          <Sider
            style={{
              position: 'fixed',
              left: 0,
              top: '56px',
              zIndex: 99,
              background: 'var(--semi-color-bg-1)',
              boxShadow: '0 2px 8px rgba(0, 0, 0, 0.15)',
              border: 'none',
              paddingRight: '0',
              height: 'calc(100vh - 56px)',
            }}
          >
            <SiderBar />
          </Sider>
        )}
        <Layout
          style={{
            marginLeft: styleState.isMobile
              ? '0'
              : styleState.showSider
                ? styleState.siderCollapsed
                  ? '60px'
                  : '200px'
                : '0',
            transition: 'margin-left 0.3s ease',
            flex: '1 1 auto',
            display: 'flex',
            flexDirection: 'column',
          }}
        >
          <Content
            style={{
              flex: '1 0 auto',
              overflowY: styleState.isMobile ? 'visible' : 'auto',
              WebkitOverflowScrolling: 'touch',
              padding: styleState.shouldInnerPadding ? '24px' : '0',
              position: 'relative',
              marginTop: styleState.isMobile ? '2px' : '0',
            }}
          >
            <App />
          </Content>
          <Layout.Footer
            style={{
              flex: '0 0 auto',
              width: '100%',
            }}
          >
            <FooterBar />
          </Layout.Footer>
        </Layout>
      </Layout>
      <ToastContainer />
    </Layout>
  );
};

export default PageLayout;
