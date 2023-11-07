import React, {useContext, useEffect, useState} from 'react';
import {Link, useNavigate} from 'react-router-dom';
import {UserContext} from '../context/User';

import {Button, Container, Icon, Menu, Segment} from 'semantic-ui-react';
import {API, getLogo, getSystemName, isAdmin, isMobile, showSuccess} from '../helpers';
import '../index.css';

import {
    IconAt,
    IconHistogram,
    IconGift,
    IconKey,
    IconUser,
    IconLayers,
    IconHelpCircle,
    IconCreditCard,
    IconSemiLogo,
    IconHome,
    IconImage
} from '@douyinfe/semi-icons';
import {Nav, Avatar, Dropdown, Layout, Switch} from '@douyinfe/semi-ui';
import {stringToColor} from "../helpers/render";

// HeaderBar Buttons
let headerButtons = [
    {
        text: '关于',
        itemKey: 'about',
        to: '/about',
        icon: <IconHelpCircle/>
    },
];

if (localStorage.getItem('chat_link')) {
    headerButtons.splice(1, 0, {
        name: '聊天',
        to: '/chat',
        icon: 'comments'
    });
}

const HeaderBar = () => {
    const [userState, userDispatch] = useContext(UserContext);
    let navigate = useNavigate();

    const [showSidebar, setShowSidebar] = useState(false);
    const [dark, setDark] = useState(false);
    const systemName = getSystemName();
    const logo = getLogo();
    var themeMode = localStorage.getItem('theme-mode');

    async function logout() {
        setShowSidebar(false);
        await API.get('/api/user/logout');
        showSuccess('注销成功!');
        userDispatch({type: 'logout'});
        localStorage.removeItem('user');
        navigate('/login');
    }

    useEffect(() => {
        if (themeMode === 'dark') {
            switchMode(true);
        }
    }, []);

    const switchMode = (model) => {
        const body = document.body;
        if (!model) {
            body.removeAttribute('theme-mode');
            localStorage.setItem('theme-mode', 'light');
        } else {
            body.setAttribute('theme-mode', 'dark');
            localStorage.setItem('theme-mode', 'dark');
        }
        setDark(model);
    };
    return (
        <>
            <Layout>
                <div style={{width: '100%'}}>
                    <Nav
                        mode={'horizontal'}
                        // bodyStyle={{ height: 100 }}
                        renderWrapper={({itemElement, isSubNav, isInSubNav, props}) => {
                            const routerMap = {
                                about: "/about",
                                login: "/login",
                                register: "/register",
                            };
                            return (
                                <Link
                                    style={{textDecoration: "none"}}
                                    to={routerMap[props.itemKey]}
                                >
                                    {itemElement}
                                </Link>
                            );
                        }}
                        selectedKeys={[]}
                        // items={headerButtons}
                        onSelect={key => {

                        }}
                        footer={
                            <>
                                <Nav.Item itemKey={'about'} icon={<IconHelpCircle />} />
                                <Switch checkedText="🌞" size={'large'} checked={dark} uncheckedText="🌙" onChange={switchMode} />
                                {userState.user ?
                                    <>
                                        <Dropdown
                                            position="bottomRight"
                                            render={
                                                <Dropdown.Menu>
                                                    <Dropdown.Item onClick={logout}>退出</Dropdown.Item>
                                                </Dropdown.Menu>
                                            }
                                        >
                                            <Avatar size="small" color={stringToColor(userState.user.username)} style={{ margin: 4 }}>
                                                {userState.user.username[0]}
                                            </Avatar>
                                            <span>{userState.user.username}</span>
                                        </Dropdown>
                                    </>
                                    :
                                    <>
                                        <Nav.Item itemKey={'login'} text={'登录'} icon={<IconKey />} />
                                        <Nav.Item itemKey={'register'} text={'注册'} icon={<IconUser />} />
                                    </>
                                }
                            </>
                        }
                    >
                    </Nav>
                </div>
            </Layout>
        </>
    );
};

export default HeaderBar;
