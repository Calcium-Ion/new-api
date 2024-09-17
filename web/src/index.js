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

// initialization

const root = ReactDOM.createRoot(document.getElementById('root'));
const { Sider, Content, Header, Footer } = Layout;
root.render(
  <React.StrictMode>
    <StatusProvider>
      <UserProvider>
        <BrowserRouter>
          <ThemeProvider>
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
                    style={{ overflowY: 'auto', padding: '24px' }}
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
          </ThemeProvider>
        </BrowserRouter>
      </UserProvider>
    </StatusProvider>
  </React.StrictMode>,
);
