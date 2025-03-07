import React, { useContext, useEffect, useMemo, useState } from 'react';
import { Link, useNavigate } from 'react-router-dom';
import { UserContext } from '../context/User';
import { StatusContext } from '../context/Status';
import { useTranslation } from 'react-i18next';

import {
  API,
  getLogo,
  getSystemName,
  isAdmin,
  isMobile,
  showError,
} from '../helpers';
import '../index.css';

import {
  IconCalendarClock, IconChecklistStroked,
  IconComment, IconCommentStroked,
  IconCreditCard,
  IconGift, IconHelpCircle,
  IconHistogram,
  IconHome,
  IconImage,
  IconKey,
  IconLayers,
  IconPriceTag,
  IconSetting,
  IconUser
} from '@douyinfe/semi-icons';
import { Avatar, Dropdown, Layout, Nav, Switch, Divider } from '@douyinfe/semi-ui';
import { setStatusData } from '../helpers/data.js';
import { stringToColor } from '../helpers/render.js';
import { useSetTheme, useTheme } from '../context/Theme/index.js';
import { StyleContext } from '../context/Style/index.js';

// HeaderBar Buttons

const SiderBar = () => {
  const { t } = useTranslation();
  const [styleState, styleDispatch] = useContext(StyleContext);
  const [statusState, statusDispatch] = useContext(StatusContext);
  const defaultIsCollapsed =
    localStorage.getItem('default_collapse_sidebar') === 'true';

  const [selectedKeys, setSelectedKeys] = useState(['home']);
  const [isCollapsed, setIsCollapsed] = useState(defaultIsCollapsed);
  const [chatItems, setChatItems] = useState([]);
  const theme = useTheme();
  const setTheme = useSetTheme();

  const routerMap = {
    home: '/',
    channel: '/channel',
    token: '/token',
    redemption: '/redemption',
    topup: '/topup',
    user: '/user',
    log: '/log',
    midjourney: '/midjourney',
    setting: '/setting',
    about: '/about',
    chat: '/chat',
    detail: '/detail',
    pricing: '/pricing',
    task: '/task',
    playground: '/playground',
    personal: '/personal',
  };

  const workspaceItems = useMemo(
    () => [
      {
        text: t('数据看板'),
        itemKey: 'detail',
        to: '/detail',
        icon: <IconCalendarClock />,
        className:
          localStorage.getItem('enable_data_export') === 'true'
            ? ''
            : 'tableHiddle',
      },
      {
        text: t('API令牌'),
        itemKey: 'token',
        to: '/token',
        icon: <IconKey />,
      },
      {
        text: t('使用日志'),
        itemKey: 'log',
        to: '/log',
        icon: <IconHistogram />,
      },
      {
        text: t('绘图日志'),
        itemKey: 'midjourney',
        to: '/midjourney',
        icon: <IconImage />,
        className:
          localStorage.getItem('enable_drawing') === 'true'
            ? ''
            : 'tableHiddle',
      },
      {
        text: t('任务日志'),
        itemKey: 'task',
        to: '/task',
        icon: <IconChecklistStroked />,
        className:
          localStorage.getItem('enable_task') === 'true'
            ? ''
            : 'tableHiddle',
      }
    ],
    [
      localStorage.getItem('enable_data_export'),
      localStorage.getItem('enable_drawing'),
      localStorage.getItem('enable_task'),
      t,
    ],
  );

  const financeItems = useMemo(
    () => [
      {
        text: t('钱包'),
        itemKey: 'topup',
        to: '/topup',
        icon: <IconCreditCard />,
      },
      {
        text: t('个人设置'),
        itemKey: 'personal',
        to: '/personal',
        icon: <IconUser />,
      },
    ],
    [t],
  );

  const adminItems = useMemo(
    () => [
      {
        text: t('渠道'),
        itemKey: 'channel',
        to: '/channel',
        icon: <IconLayers />,
        className: isAdmin() ? '' : 'tableHiddle',
      },
      {
        text: t('兑换码'),
        itemKey: 'redemption',
        to: '/redemption',
        icon: <IconGift />,
        className: isAdmin() ? '' : 'tableHiddle',
      },
      {
        text: t('用户管理'),
        itemKey: 'user',
        to: '/user',
        icon: <IconUser />,
      },
      {
        text: t('系统设置'),
        itemKey: 'setting',
        to: '/setting',
        icon: <IconSetting />,
      },
    ],
    [isAdmin(), t],
  );

  const chatMenuItems = useMemo(
    () => [
      {
        text: 'Playground',
        itemKey: 'playground',
        to: '/playground',
        icon: <IconCommentStroked />,
      },
      {
        text: t('聊天'),
        itemKey: 'chat',
        items: chatItems,
        icon: <IconComment />,
      },
    ],
    [chatItems, t],
  );

  useEffect(() => {
    let localKey = window.location.pathname.split('/')[1];
    if (localKey === '') {
      localKey = 'home';
    }
    setSelectedKeys([localKey]);

    let chatLink = localStorage.getItem('chat_link');
    if (!chatLink) {
      let chats = localStorage.getItem('chats');
      if (chats) {
        // console.log(chats);
        try {
          chats = JSON.parse(chats);
          if (Array.isArray(chats)) {
            let chatItems = [];
            for (let i = 0; i < chats.length; i++) {
              let chat = {};
              for (let key in chats[i]) {
                chat.text = key;
                chat.itemKey = 'chat' + i;
                chat.to = '/chat/' + i;
              }
              // setRouterMap({ ...routerMap, chat: '/chat/' + i })
              chatItems.push(chat);
            }
            setChatItems(chatItems);
          }
        } catch (e) {
          console.error(e);
          showError('聊天数据解析失败')
        }
      }
    }

    setIsCollapsed(localStorage.getItem('default_collapse_sidebar') === 'true');
  }, []);

  // Custom divider style
  const dividerStyle = {
    margin: '8px 0',
    opacity: 0.6,
  };

  // Custom group label style
  const groupLabelStyle = {
    padding: '8px 16px',
    color: 'var(--semi-color-text-2)',
    fontSize: '12px',
    fontWeight: 'normal',
  };

  return (
    <>
      <Nav
        style={{ maxWidth: 200, height: '100%' }}
        defaultIsCollapsed={
          localStorage.getItem('default_collapse_sidebar') === 'true'
        }
        isCollapsed={isCollapsed}
        onCollapseChange={(collapsed) => {
          setIsCollapsed(collapsed);
        }}
        selectedKeys={selectedKeys}
        renderWrapper={({ itemElement, isSubNav, isInSubNav, props }) => {
          let chatLink = localStorage.getItem('chat_link');
          if (!chatLink) {
            let chats = localStorage.getItem('chats');
            if (chats) {
              chats = JSON.parse(chats);
              if (Array.isArray(chats) && chats.length > 0) {
                for (let i = 0; i < chats.length; i++) {
                  routerMap['chat' + i] = '/chat/' + i;
                }
                if (chats.length > 1) {
                  // delete /chat
                  if (routerMap['chat']) {
                    delete routerMap['chat'];
                  }
                } else {
                  // rename /chat to /chat/0
                  routerMap['chat'] = '/chat/0';
                }
              }
            }
          }
          return (
            <Link
              style={{ textDecoration: 'none' }}
              to={routerMap[props.itemKey]}
            >
              {itemElement}
            </Link>
          );
        }}
        onSelect={(key) => {
          if (key.itemKey.toString().startsWith('chat')) {
            styleDispatch({ type: 'SET_INNER_PADDING', payload: false });
          } else {
            styleDispatch({ type: 'SET_INNER_PADDING', payload: true });
          }
          setSelectedKeys([key.itemKey]);
        }}
      >
        {/* Chat Section - Only show if there are chat items */}
        {chatItems.length > 0 && (
          <>
            {chatMenuItems.map((item) => {
              if (item.items && item.items.length > 0) {
                return (
                  <Nav.Sub
                    key={item.itemKey}
                    itemKey={item.itemKey}
                    text={item.text}
                    icon={item.icon}
                  >
                    {item.items.map((subItem) => (
                      <Nav.Item
                        key={subItem.itemKey}
                        itemKey={subItem.itemKey}
                        text={subItem.text}
                      />
                    ))}
                  </Nav.Sub>
                );
              } else {
                return (
                  <Nav.Item
                    key={item.itemKey}
                    itemKey={item.itemKey}
                    text={item.text}
                    icon={item.icon}
                  />
                );
              }
            })}
          </>
        )}

        {/* Divider */}
        <Divider style={dividerStyle} />

        {/* Workspace Section */}
        {!isCollapsed && <div style={groupLabelStyle}>{t('控制台')}</div>}
        {workspaceItems.map((item) => (
          <Nav.Item
            key={item.itemKey}
            itemKey={item.itemKey}
            text={item.text}
            icon={item.icon}
            className={item.className}
          />
        ))}

        {/* Divider */}
        <Divider style={dividerStyle} />

        {/* Finance Management Section */}
        {!isCollapsed && <div style={groupLabelStyle}>{t('个人中心')}</div>}
        {financeItems.map((item) => (
          <Nav.Item
            key={item.itemKey}
            itemKey={item.itemKey}
            text={item.text}
            icon={item.icon}
            className={item.className}
          />
        ))}

        {isAdmin() && (
          <>
            {/* Divider */}
            <Divider style={dividerStyle} />

            {/* Admin Section */}
            {adminItems.map((item) => (
              <Nav.Item
                key={item.itemKey}
                itemKey={item.itemKey}
                text={item.text}
                icon={item.icon}
                className={item.className}
              />
            ))}
          </>
        )}

        <Nav.Footer
          collapseButton={true}
          collapseText={(collapsed)=>
            {
              if(collapsed){
                return t('展开侧边栏')
              }
                return t('收起侧边栏')
            }
          }
        />
      </Nav>
    </>
  );
};

export default SiderBar;
