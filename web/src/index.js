import React from 'react';
import ReactDOM from 'react-dom/client';
import { BrowserRouter } from 'react-router-dom';
import App from './App';
import HeaderBar from './components/HeaderBar';
import 'semantic-ui-offline/semantic.min.css';
import './index.css';
import { UserProvider } from './context/User';
import { ToastContainer } from 'react-toastify';
import 'react-toastify/dist/ReactToastify.css';
import { StatusProvider } from './context/Status';
import { Layout } from '@douyinfe/semi-ui';
import SiderBar from './components/SiderBar';
import { ThemeProvider } from './context/Theme';
import FooterBar from './components/Footer';
import { PageProvider, usePageContext } from './contexts/PageContext';

// initialization

const root = ReactDOM.createRoot(document.getElementById('root'));
const { Sider, Content, Header, Footer } = Layout;

const AppLayout = () => {
  // ChatPage 组件下 isChat 设置为 true，使得进入聊天页面时，Content 的 padding 将设置为 0px，并且 Footer 将被隐藏。而在其他页面，将保持原有的设置。
  const { isChat } = usePageContext();

  return (
    <Layout style={{ height: '100vh', display: 'flex', flexDirection: 'column' }}>
      <Header>
        <HeaderBar />
      </Header>
      <Layout style={{ flex: 1, overflow: 'hidden' }}>
        <Sider>
          <SiderBar />
        </Sider>
        <Layout>
          <Content
            style={{ 
              overflowY: 'auto', 
              padding: isChat ? '0px' : '24px',
              height: '100%'
            }}
          >
            <App />
          </Content>
          {!isChat && (
            <Layout.Footer
              style={{
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                height: '64px', // 您可以根据需要调整这个高度
              }}
            >
              <FooterBar />
            </Layout.Footer>
          )}
        </Layout>
      </Layout>
      <ToastContainer />
    </Layout>
  );
};

root.render(
  <React.StrictMode>
    <StatusProvider>
      <UserProvider>
        <BrowserRouter>
          <ThemeProvider>
            <PageProvider>
              <AppLayout />
            </PageProvider>
          </ThemeProvider>
        </BrowserRouter>
      </UserProvider>
    </StatusProvider>
  </React.StrictMode>,
);
