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
import PasswordResetConfirm from './components/PasswordResetConfirm';
import { UserContext } from './context/User';
import Channel from './pages/Channel';
import Token from './pages/Token';
import EditChannel from './pages/Channel/EditChannel';
import Redemption from './pages/Redemption';
import TopUp from './pages/TopUp';
import Log from './pages/Log';
import Chat from './pages/Chat';
import Chat2Link from './pages/Chat2Link';
import { Layout } from '@douyinfe/semi-ui';
import Midjourney from './pages/Midjourney';
import Pricing from './pages/Pricing/index.js';
import Task from "./pages/Task/index.js";
import Playground from './pages/Playground/Playground.js';
import OAuth2Callback from "./components/OAuth2Callback.js";
import { useTranslation } from 'react-i18next';
import { StatusContext } from './context/Status';
import { setStatusData } from './helpers/data.js';
import { API, showError } from './helpers';

const Home = lazy(() => import('./pages/Home'));
const Detail = lazy(() => import('./pages/Detail'));
const About = lazy(() => import('./pages/About'));

function App() {
  return (
    <>
      <Routes>
        <Route
          path='/'
          element={
            <Suspense fallback={<Loading></Loading>}>
              <Home />
            </Suspense>
          }
        />
        <Route
          path='/channel'
          element={
            <PrivateRoute>
              <Channel />
            </PrivateRoute>
          }
        />
        <Route
          path='/channel/edit/:id'
          element={
            <Suspense fallback={<Loading></Loading>}>
              <EditChannel />
            </Suspense>
          }
        />
        <Route
          path='/channel/add'
          element={
            <Suspense fallback={<Loading></Loading>}>
              <EditChannel />
            </Suspense>
          }
        />
        <Route
          path='/token'
          element={
            <PrivateRoute>
              <Token />
            </PrivateRoute>
          }
        />
        <Route
          path='/playground'
          element={
            <PrivateRoute>
              <Playground />
            </PrivateRoute>
          }
        />
        <Route
          path='/redemption'
          element={
            <PrivateRoute>
              <Redemption />
            </PrivateRoute>
          }
        />
        <Route
          path='/user'
          element={
            <PrivateRoute>
              <User />
            </PrivateRoute>
          }
        />
        <Route
          path='/user/edit/:id'
          element={
            <Suspense fallback={<Loading></Loading>}>
              <EditUser />
            </Suspense>
          }
        />
        <Route
          path='/user/edit'
          element={
            <Suspense fallback={<Loading></Loading>}>
              <EditUser />
            </Suspense>
          }
        />
        <Route
          path='/user/reset'
          element={
            <Suspense fallback={<Loading></Loading>}>
              <PasswordResetConfirm />
            </Suspense>
          }
        />
        <Route
          path='/login'
          element={
            <Suspense fallback={<Loading></Loading>}>
              <LoginForm />
            </Suspense>
          }
        />
        <Route
          path='/register'
          element={
            <Suspense fallback={<Loading></Loading>}>
              <RegisterForm />
            </Suspense>
          }
        />
        <Route
          path='/reset'
          element={
            <Suspense fallback={<Loading></Loading>}>
              <PasswordResetForm />
            </Suspense>
          }
        />
        <Route
          path='/oauth/github'
          element={
            <Suspense fallback={<Loading></Loading>}>
              <OAuth2Callback type='github'></OAuth2Callback>
            </Suspense>
          }
        />
        <Route
          path='/oauth/linuxdo'
          element={
            <Suspense fallback={<Loading></Loading>}>
                <OAuth2Callback type='linuxdo'></OAuth2Callback>
            </Suspense>
          }
        />
        <Route
          path='/setting'
          element={
            <PrivateRoute>
              <Suspense fallback={<Loading></Loading>}>
                <Setting />
              </Suspense>
            </PrivateRoute>
          }
        />
        <Route
          path='/topup'
          element={
            <PrivateRoute>
              <Suspense fallback={<Loading></Loading>}>
                <TopUp />
              </Suspense>
            </PrivateRoute>
          }
        />
        <Route
          path='/log'
          element={
            <PrivateRoute>
              <Log />
            </PrivateRoute>
          }
        />
        <Route
          path='/detail'
          element={
            <PrivateRoute>
              <Suspense fallback={<Loading></Loading>}>
                <Detail />
              </Suspense>
            </PrivateRoute>
          }
        />
        <Route
          path='/midjourney'
          element={
            <PrivateRoute>
              <Suspense fallback={<Loading></Loading>}>
                <Midjourney />
              </Suspense>
            </PrivateRoute>
          }
        />
        <Route
          path='/task'
          element={
            <PrivateRoute>
              <Suspense fallback={<Loading></Loading>}>
                <Task />
              </Suspense>
            </PrivateRoute>
          }
        />
        <Route
          path='/pricing'
          element={
            <Suspense fallback={<Loading></Loading>}>
              <Pricing />
            </Suspense>
          }
        />
        <Route
          path='/about'
          element={
            <Suspense fallback={<Loading></Loading>}>
              <About />
            </Suspense>
          }
        />
        <Route
          path='/chat/:id?'
          element={
            <Suspense fallback={<Loading></Loading>}>
              <Chat />
            </Suspense>
          }
        />
        {/* 方便使用chat2link直接跳转聊天... */}
          <Route
            path='/chat2link'
            element={
              <PrivateRoute>
                <Suspense fallback={<Loading></Loading>}>
                    <Chat2Link />
                </Suspense>
              </PrivateRoute>
            }
          />
          <Route path='*' element={<NotFound />} />
        </Routes>
      </>
  );
}

export default App;
