import React from 'react';
import ReactDOM from 'react-dom/client';
import { BrowserRouter } from 'react-router-dom';
import App from './App';
import HeaderBar from './components/HeaderBar';
import Footer from './components/Footer';
import 'semantic-ui-offline/semantic.min.css';
import './index.css';
import { UserProvider } from './context/User';
import { ToastContainer } from 'react-toastify';
import 'react-toastify/dist/ReactToastify.css';
import { StatusProvider } from './context/Status';
import { Layout } from '@douyinfe/semi-ui';
import SiderBar from './components/SiderBar';
import { ThemeProvider } from './context/Theme';

// initialization

const root = ReactDOM.createRoot(document.getElementById('root'));
const { Sider, Content, Header } = Layout;
root.render(
  <React.StrictMode>
    <StatusProvider>
      <UserProvider>
        <BrowserRouter>
          <ThemeProvider>
            <Layout>
              <Sider>
                <SiderBar />
              </Sider>
              <Layout>
                <Header>
                  <HeaderBar />
                </Header>
                <Content
                  style={{
                    padding: '24px',
                  }}
                >
                  <App />
                </Content>
                <Layout.Footer>
                  <Footer></Footer>
                </Layout.Footer>
              </Layout>
              <ToastContainer />
            </Layout>
          </ThemeProvider>
        </BrowserRouter>
      </UserProvider>
    </StatusProvider>
  </React.StrictMode>,
);
