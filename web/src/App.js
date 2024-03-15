import React, { lazy, Suspense, useContext, useEffect } from 'react';
import { Route, Routes } from 'react-router-dom';
import Loading from './components/Loading';
import User from './pages/User';
import { PrivateRoute } from './components/PrivateRoute';
import RegisterForm from './components/RegisterForm';
import LoginForm from './components/LoginForm';
import NotFound from './pages/NotFound';
import Setting from './pages/Setting';
import EditUser from './pages/User/EditUser';
import { getLogo, getSystemName } from './helpers';
import PasswordResetForm from './components/PasswordResetForm';
import GitHubOAuth from './components/GitHubOAuth';
import PasswordResetConfirm from './components/PasswordResetConfirm';
import { UserContext } from './context/User';
import Channel from './pages/Channel';
import Token from './pages/Token';
import EditChannel from './pages/Channel/EditChannel';
import Redemption from './pages/Redemption';
import TopUp from './pages/TopUp';
import Log from './pages/Log';
import Chat from './pages/Chat';
import { Layout } from '@douyinfe/semi-ui';
import Midjourney from './pages/Midjourney';
import Detail from './pages/Detail';

const Home = lazy(() => import('./pages/Home'));
const About = lazy(() => import('./pages/About'));

function App() {
  const [userState, userDispatch] = useContext(UserContext);
  // const [statusState, statusDispatch] = useContext(StatusContext);

  const loadUser = () => {
    let user = localStorage.getItem('user');
    if (user) {
      let data = JSON.parse(user);
      userDispatch({ type: 'login', payload: data });
    }
  };

  useEffect(() => {
    loadUser();
    let systemName = getSystemName();
    if (systemName) {
      document.title = systemName;
    }
    let logo = getLogo();
    if (logo) {
      let linkElement = document.querySelector('link[rel~=\'icon\']');
      if (linkElement) {
        linkElement.href = logo;
      }
    }
  }, []);

  return (
    <Layout>
      <Layout.Content>
        <Routes>
          <Route
            path="/"
            element={
              <Suspense fallback={<Loading></Loading>}>
                <Home />
              </Suspense>
            }
          />
          <Route
            path="/channel"
            element={
              <PrivateRoute>
                <Channel />
              </PrivateRoute>
            }
          />
          <Route
            path="/channel/edit/:id"
            element={
              <Suspense fallback={<Loading></Loading>}>
                <EditChannel />
              </Suspense>
            }
          />
          <Route
            path="/channel/add"
            element={
              <Suspense fallback={<Loading></Loading>}>
                <EditChannel />
              </Suspense>
            }
          />
          <Route
            path="/token"
            element={
              <PrivateRoute>
                <Token />
              </PrivateRoute>
            }
          />
          <Route
            path="/redemption"
            element={
              <PrivateRoute>
                <Redemption />
              </PrivateRoute>
            }
          />
          <Route
            path="/user"
            element={
              <PrivateRoute>
                <User />
              </PrivateRoute>
            }
          />
          <Route
            path="/user/edit/:id"
            element={
              <Suspense fallback={<Loading></Loading>}>
                <EditUser />
              </Suspense>
            }
          />
          <Route
            path="/user/edit"
            element={
              <Suspense fallback={<Loading></Loading>}>
                <EditUser />
              </Suspense>
            }
          />
          <Route
            path="/user/reset"
            element={
              <Suspense fallback={<Loading></Loading>}>
                <PasswordResetConfirm />
              </Suspense>
            }
          />
          <Route
            path="/login"
            element={
              <Suspense fallback={<Loading></Loading>}>
                <LoginForm />
              </Suspense>
            }
          />
          <Route
            path="/register"
            element={
              <Suspense fallback={<Loading></Loading>}>
                <RegisterForm />
              </Suspense>
            }
          />
          <Route
            path="/reset"
            element={
              <Suspense fallback={<Loading></Loading>}>
                <PasswordResetForm />
              </Suspense>
            }
          />
          <Route
            path="/oauth/github"
            element={
              <Suspense fallback={<Loading></Loading>}>
                <GitHubOAuth />
              </Suspense>
            }
          />
          <Route
            path="/setting"
            element={
              <PrivateRoute>
                <Suspense fallback={<Loading></Loading>}>
                  <Setting />
                </Suspense>
              </PrivateRoute>
            }
          />
          <Route
            path="/topup"
            element={
              <PrivateRoute>
                <Suspense fallback={<Loading></Loading>}>
                  <TopUp />
                </Suspense>
              </PrivateRoute>
            }
          />
          <Route
            path="/log"
            element={
              <PrivateRoute>
                <Log />
              </PrivateRoute>
            }
          />
          <Route
            path="/detail"
            element={
              <PrivateRoute>
                <Detail />
              </PrivateRoute>
            }
          />
          <Route
            path="/midjourney"
            element={
              <PrivateRoute>
                <Midjourney />
              </PrivateRoute>
            }
          />
          <Route
            path="/about"
            element={
              <Suspense fallback={<Loading></Loading>}>
                <About />
              </Suspense>
            }
          />
          <Route
            path="/chat"
            element={
              <Suspense fallback={<Loading></Loading>}>
                <Chat />
              </Suspense>
            }
          />
          <Route path="*" element={
            <NotFound />
          } />
        </Routes>
      </Layout.Content>
    </Layout>
  );
}

export default App;
