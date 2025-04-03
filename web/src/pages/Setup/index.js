import React, { useContext, useEffect, useState, useRef } from 'react';
import { Card, Col, Row, Form, Button, Typography, Space, RadioGroup, Radio, Modal, Banner } from '@douyinfe/semi-ui';
import { API, showError, showNotice, timestamp2string } from '../../helpers';
import { StatusContext } from '../../context/Status';
import { marked } from 'marked';
import { StyleContext } from '../../context/Style/index.js';
import { useTranslation } from 'react-i18next';
import { IconHelpCircle, IconInfoCircle, IconAlertTriangle } from '@douyinfe/semi-icons';

const Setup = () => {
  const { t, i18n } = useTranslation();
  const [statusState] = useContext(StatusContext);
  const [styleState, styleDispatch] = useContext(StyleContext);
  const [loading, setLoading] = useState(false);
  const [selfUseModeInfoVisible, setUsageModeInfoVisible] = useState(false);
  const [setupStatus, setSetupStatus] = useState({
    status: false,
    root_init: false,
    database_type: ''
  });
  const { Text, Title } = Typography;
  const formRef = useRef(null);
  
  const [formData, setFormData] = useState({
    username: '',
    password: '',
    confirmPassword: '',
    usageMode: 'external'
  });

  useEffect(() => {
    fetchSetupStatus();
  }, []);

  const fetchSetupStatus = async () => {
    try {
      const res = await API.get('/api/setup');
      const { success, data } = res.data;
      if (success) {
        setSetupStatus(data);
        
        // If setup is already completed, redirect to home
        if (data.status) {
          window.location.href = '/';
        }
      } else {
        showError(t('获取初始化状态失败'));
      }
    } catch (error) {
      console.error('Failed to fetch setup status:', error);
      showError(t('获取初始化状态失败'));
    }
  };

  const handleUsageModeChange = (val) => {
    setFormData({...formData, usageMode: val});
  };

  const onSubmit = () => {
    if (!formRef.current) {
      console.error("Form reference is null");
      showError(t('表单引用错误，请刷新页面重试'));
      return;
    }
    
    const values = formRef.current.getValues();
    console.log("Form values:", values);
    
    // For root_init=false, validate admin username and password
    if (!setupStatus.root_init) {
      if (!values.username || !values.username.trim()) {
        showError(t('请输入管理员用户名'));
        return;
      }
      
      if (!values.password || values.password.length < 8) {
        showError(t('密码长度至少为8个字符'));
        return;
      }
      
      if (values.password !== values.confirmPassword) {
        showError(t('两次输入的密码不一致'));
        return;
      }
    }
    
    // Prepare submission data
    const formValues = {...values};
    formValues.SelfUseModeEnabled = values.usageMode === 'self';
    formValues.DemoSiteEnabled = values.usageMode === 'demo';
    
    // Remove usageMode as it's not needed by the backend
    delete formValues.usageMode;
    
    console.log("Submitting data to backend:", formValues);
    setLoading(true);
    
    // Submit to backend
    API.post('/api/setup', formValues)
      .then(res => {
        const { success, message } = res.data;
        console.log("API response:", res.data);
        
        if (success) {
          showNotice(t('系统初始化成功，正在跳转...'));
          setTimeout(() => {
            window.location.reload();
          }, 1500);
        } else {
          showError(message || t('初始化失败，请重试'));
        }
      })
      .catch(error => {
        console.error('API error:', error);
        showError(t('系统初始化失败，请重试'));
        setLoading(false);
      })
      .finally(() => {
        // setLoading(false);
      });
  };

  return (
    <>
      <div style={{ maxWidth: '800px', margin: '0 auto', padding: '20px' }}>
        <Card>
          <Title heading={2} style={{ marginBottom: '24px' }}>{t('系统初始化')}</Title>
          
          {setupStatus.database_type === 'sqlite' && (
            <Banner
              type="warning"
              icon={<IconAlertTriangle size="large" />}
              closeIcon={null}
              title={t('数据库警告')}
              description={
                <div>
                  <p>{t('您正在使用 SQLite 数据库。如果您在容器环境中运行，请确保已正确设置数据库文件的持久化映射，否则容器重启后所有数据将丢失！')}</p>
                  <p>{t('建议在生产环境中使用 MySQL 或 PostgreSQL 数据库，或确保 SQLite 数据库文件已映射到宿主机的持久化存储。')}</p>
                </div>
              }
              style={{ marginBottom: '24px' }}
            />
          )}
          
          <Form
            getFormApi={(formApi) => { formRef.current = formApi; console.log("Form API set:", formApi); }}
            initValues={formData}
          >
            {setupStatus.root_init ? (
              <Banner
                type="info"
                icon={<IconInfoCircle />}
                closeIcon={null}
                description={t('管理员账号已经初始化过，请继续设置系统参数')}
                style={{ marginBottom: '24px' }}
              />
            ) : (
              <Form.Section text={t('管理员账号')}>
                <Form.Input
                  field="username"
                  label={t('用户名')}
                  placeholder={t('请输入管理员用户名')}
                  showClear
                  onChange={(value) => setFormData({...formData, username: value})}
                />
                <Form.Input
                  field="password"
                  label={t('密码')}
                  placeholder={t('请输入管理员密码')}
                  type="password"
                  showClear
                  onChange={(value) => setFormData({...formData, password: value})}
                />
                <Form.Input
                  field="confirmPassword"
                  label={t('确认密码')}
                  placeholder={t('请确认管理员密码')}
                  type="password"
                  showClear
                  onChange={(value) => setFormData({...formData, confirmPassword: value})}
                />
              </Form.Section>
            )}
            
            <Form.Section text={
              <div style={{ display: 'flex', alignItems: 'center' }}>
                {t('系统设置')}
              </div>
            }>
              <Form.RadioGroup 
                field="usageMode" 
                label={
                  <div style={{ display: 'flex', alignItems: 'center' }}>
                    {t('使用模式')}
                    <IconHelpCircle 
                      style={{ marginLeft: '4px', color: 'var(--semi-color-primary)', verticalAlign: 'middle', cursor: 'pointer' }} 
                      onClick={(e) => {
                        // e.preventDefault();
                        // e.stopPropagation();
                        setUsageModeInfoVisible(true);
                      }}
                    />
                  </div>
                }
                extraText={t('可在初始化后修改')}
                initValue="external"
                onChange={handleUsageModeChange}
              >
                <Form.Radio value="external">{t('对外运营模式')}</Form.Radio>
                <Form.Radio value="self">{t('自用模式')}</Form.Radio>
                <Form.Radio value="demo">{t('演示站点模式')}</Form.Radio>
              </Form.RadioGroup>
            </Form.Section>
          </Form>

          <div style={{ marginTop: '24px', textAlign: 'right' }}>
            <Button type="primary" onClick={onSubmit} loading={loading}>
              {t('初始化系统')}
            </Button>
          </div>
        </Card>
      </div>

      <Modal
        title={t('使用模式说明')}
        visible={selfUseModeInfoVisible}
        onOk={() => setUsageModeInfoVisible(false)}
        onCancel={() => setUsageModeInfoVisible(false)}
        closeOnEsc={true}
        okText={t('确定')}
        cancelText={null}
      >
        <div style={{ padding: '8px 0' }}>
          <Title heading={6}>{t('对外运营模式')}</Title>
          <p>{t('默认模式，适用于为多个用户提供服务的场景。')}</p>
          <p>{t('此模式下，系统将计算每次调用的用量，您需要对每个模型都设置价格，如果没有设置价格，用户将无法使用该模型。')}</p>
        </div>
        <div style={{ padding: '8px 0' }}>
          <Title heading={6}>{t('自用模式')}</Title>
          <p>{t('适用于个人使用的场景。')}</p>
          <p>{t('不需要设置模型价格，系统将弱化用量计算，您可专注于使用模型。')}</p>
        </div>
        <div style={{ padding: '8px 0' }}>
          <Title heading={6}>{t('演示站点模式')}</Title>
          <p>{t('适用于展示系统功能的场景。')}</p>
        </div>
      </Modal>
    </>
  );
};

export default Setup;
