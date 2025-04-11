import React, { useEffect, useState } from 'react';
import { API, isMobile, showError, showInfo, showSuccess } from '../../helpers';
import {
  renderNumber,
  renderQuota,
  renderQuotaWithAmount,
} from '../../helpers/render';
import {
  Col,
  Layout,
  Row,
  Typography,
  Card,
  Button,
  Form,
  Divider,
  Space,
  Modal,
  Toast,
} from '@douyinfe/semi-ui';
import Title from '@douyinfe/semi-ui/lib/es/typography/title';
import Text from '@douyinfe/semi-ui/lib/es/typography/text';
import { Link } from 'react-router-dom';
import { useTranslation } from 'react-i18next';

const TopUp = () => {
  const { t } = useTranslation();
  const [redemptionCode, setRedemptionCode] = useState('');
  const [topUpCode, setTopUpCode] = useState('');
  const [topUpCount, setTopUpCount] = useState(0);
  const [minTopupCount, setMinTopUpCount] = useState(1);
  const [amount, setAmount] = useState(0.0);
  const [minTopUp, setMinTopUp] = useState(1);
  const [stripeTopUpCount, setStripeTopUpCount] = useState(0);
  const [stripeAmount, setStripeAmount] = useState(0.0);
  const [stripeMinTopUp, setStripeMinTopUp] = useState(1);
  const [topUpLink, setTopUpLink] = useState('');
  const [enableOnlineTopUp, setEnableOnlineTopUp] = useState(false);
  const [enableStripeTopUp, setEnableStripeTopUp] = useState(false);
  const [userQuota, setUserQuota] = useState(0);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [open, setOpen] = useState(false);
  const [payWay, setPayWay] = useState('');
  const [isPaying, setIsPaying] = useState(false);

  const topUp = async () => {
    if (redemptionCode === '') {
      showInfo(t('请输入兑换码！'));
      return;
    }
    setIsSubmitting(true);
    try {
      const res = await API.post('/api/user/topup', {
        key: redemptionCode,
      });
      const { success, message, data } = res.data;
      if (success) {
        showSuccess(t('兑换成功！'));
        Modal.success({
          title: t('兑换成功！'),
          content: t('成功兑换额度：') + renderQuota(data),
          centered: true,
        });
        setUserQuota((quota) => {
          return quota + data;
        });
        setRedemptionCode('');
      } else {
        showError(message);
      }
    } catch (err) {
      showError(t('请求失败'));
    } finally {
      setIsSubmitting(false);
    }
  };

  const openTopUpLink = () => {
    if (!topUpLink) {
      showError(t('超级管理员未设置充值链接！'));
      return;
    }
    window.open(topUpLink, '_blank');
  };

  const preTopUp = async (payment) => {
    if (((payment === "zfb" || payment === "wx") && !enableOnlineTopUp) || (payment === "stripe"  && !enableStripeTopUp)) {
      showError(t('管理员未开启在线充值！'));
      return;
    }
    await getAmount(payment);
    if (!checkMinTopUp(payment)) {
      return;
    }
    setPayWay(payment);
    setOpen(true);
  };

  const onlineTopUp = async () => {
    if (amount === 0) {
      await getAmount(payWay);
    }
    if (!checkMinTopUp(payWay)) {
      return;
    }
    setOpen(false);
    try {
      setIsPaying(true);
      const res = await API.post('/api/user/pay', {
        amount: parseInt(getTopUpCount()),
        top_up_code: topUpCode,
        payment_method: payWay,
      });
      if (res !== undefined) {
        const { message, data } = res.data;
        // showInfo(message);
        if (message === 'success') {
          switch (payWay) {
            case "zfb":
            case "wx":
              processEpayCallback(data)
              break
            case "stripe":
              processStripeCallback(data)
              break
          }
        } else {
          setIsPaying(false);
          showError(data);
          // setTopUpCount(parseInt(res.data.count));
          // setAmount(parseInt(data));
        }
      } else {
        setIsPaying(false);
        showError(res);
      }
    } catch (err) {
      console.log(err);
    } finally {
    }
  };

  const processEpayCallback = (data) => {
    let params = data.params;
    let url = data.url;
    let form = document.createElement('form');
    form.action = url;
    form.method = 'POST';
    // 判断是否为safari浏览器
    let isSafari =
        navigator.userAgent.indexOf('Safari') > -1 &&
        navigator.userAgent.indexOf('Chrome') < 1;
    if (!isSafari) {
      form.target = '_blank';
    }
    for (let key in params) {
      let input = document.createElement('input');
      input.type = 'hidden';
      input.name = key;
      input.value = params[key];
      form.appendChild(input);
    }
    document.body.appendChild(form);
    form.submit();
    document.body.removeChild(form);
  }

  const processStripeCallback = (data) => {
    location.href = data.pay_link;
  };

  const checkMinTopUp = (method) => {
    let localMinTopUp
    let value
    switch (method) {
      case "zfb":
      case "wx":
        localMinTopUp = minTopUp
        value = topUpCount
        break
      case "stripe":
        localMinTopUp = stripeMinTopUp
        value = stripeTopUpCount
        break
      default:
        showError("错误的支付渠道")
        return false
    }

    if (value < localMinTopUp) {
      showError(t('充值数量不能小于') + localMinTopUp);
      return false
    }
    return true
  }

  const getTopUpCount = (method) => {
    if (method === undefined) {
      method = payWay
    }
    switch (method) {
      case "zfb":
      case "wx":
        return topUpCount
      case "stripe":
        return stripeTopUpCount
    }
  }

  const getUserQuota = async () => {
    let res = await API.get(`/api/user/self`);
    const { success, message, data } = res.data;
    if (success) {
      setUserQuota(data.quota);
    } else {
      showError(message);
    }
  };

  useEffect(() => {
    let status = localStorage.getItem('status');
    if (status) {
      status = JSON.parse(status);
      if (status.top_up_link) {
        setTopUpLink(status.top_up_link);
      }
      if (status.min_topup) {
        setMinTopUp(status.min_topup);
      }
      if (status.stripe_min_topup) {
        setStripeMinTopUp(status.stripe_min_topup)
      }
      if (status.enable_online_topup) {
        setEnableOnlineTopUp(status.enable_online_topup);
      }
      if (status.enable_stripe_topup) {
        setEnableStripeTopUp(status.enable_stripe_topup)
      }
    }
    getUserQuota().then();
  }, []);

  const renderAmountByMethod = () => {
    switch (payWay) {
      case "zfb":
      case "wx":
        return renderAmount()
      case "stripe":
        return renderStripeAmount()
      default:
        return 0
    }
  }

  const renderAmount = () => {
    // console.log(amount);
    return amount + ' ' + t('元');
  };

  const renderStripeAmount = () => {
    // console.log(amount);
    return stripeAmount + '元';
  };

  const getAmount = async (method, value) => {
    if (method === undefined) {
      showError("错误的支付渠道")
      return
    }
    if (value === undefined) {
      value = getTopUpCount(method)
    }
    try {
      const res = await API.post('/api/user/amount', {
        amount: parseFloat(value),
        top_up_code: topUpCode,
        payment_method: method,
      });
      if (res !== undefined) {
        const { message, data } = res.data;
        // showInfo(message);
        if (message === 'success') {
          setAmountByMethod(method, parseFloat(data))
        } else {
          setAmountByMethod(method, 0)
          Toast.error({ content: '错误：' + data, id: 'getAmount' });
          // setTopUpCount(parseInt(res.data.count));
          // setAmount(parseInt(data));
        }
      } else {
        showError(res);
      }
    } catch (err) {
      console.log(err);
    } finally {
    }
  };

  const setAmountByMethod = (method, value) => {
    switch (method) {
      case "zfb":
      case "wx":
        setAmount(value);
        break
      case "stripe":
        setStripeAmount(value)
        break
    }
  }

  const handleCancel = () => {
    setOpen(false);
  };

  return (
    <div>
      <Layout>
        <Layout.Header>
          <h3>{t('我的钱包')}</h3>
        </Layout.Header>
        <Layout.Content>
          <Modal
            title={t('确定要充值吗')}
            visible={open}
            onOk={onlineTopUp}
            onCancel={handleCancel}
            maskClosable={false}
            size={'small'}
            centered={true}
          >
            <p>{t('充值数量')}：{getTopUpCount()}</p>
            <p>{t('实付金额')}：{renderAmountByMethod()}</p>
            <p>{t('是否确认充值？')}</p>
          </Modal>
          <div
            style={{ marginTop: 20, display: 'flex', justifyContent: 'center' }}
          >
            <Card style={{ width: '500px', padding: '20px' }}>
              <Title level={3} style={{ textAlign: 'center' }}>
                {t('余额')} {renderQuota(userQuota)}
              </Title>
              <div style={{ marginTop: 20 }}>
                <Divider>{t('兑换余额')}</Divider>
                <Form>
                  <Form.Input
                    field={'redemptionCode'}
                    label={t('兑换码')}
                    placeholder={t('兑换码')}
                    name='redemptionCode'
                    value={redemptionCode}
                    onChange={(value) => {
                      setRedemptionCode(value);
                    }}
                  />
                  <Space>
                    {topUpLink ? (
                      <Button
                        type={'primary'}
                        theme={'solid'}
                        onClick={openTopUpLink}
                      >
                        {t('获取兑换码')}
                      </Button>
                    ) : null}
                    <Button
                      type={'warning'}
                      theme={'solid'}
                      onClick={topUp}
                      disabled={isSubmitting}
                    >
                      {isSubmitting ? t('兑换中...') : t('兑换')}
                    </Button>
                  </Space>
                </Form>
              </div>
              <div style={{ marginTop: 20 }}>
                <Divider>{t('在线充值')}</Divider>
                <Form>
                  <Form.Input
                    disabled={!enableOnlineTopUp}
                    field={'redemptionCount'}
                    label={t('实付金额：') + ' ' + renderAmount()}
                    placeholder={t('充值数量，最低 ') + renderQuotaWithAmount(minTopUp)}
                    name='redemptionCount'
                    type={'number'}
                    value={topUpCount}
                    onChange={async (value) => {
                      if (value < 1) {
                        value = 1;
                      }
                      setTopUpCount(value);
                      await getAmount("zfb", value);
                    }}
                  />
                  <Space>
                    <Button
                      type={'primary'}
                      theme={'solid'}
                      onClick={async () => {
                        preTopUp('zfb');
                      }}
                    >
                      {t('支付宝')}
                    </Button>
                    <Button
                      style={{
                        backgroundColor: 'rgba(var(--semi-green-5), 1)',
                      }}
                      type={'primary'}
                      theme={'solid'}
                      onClick={async () => {
                        preTopUp('wx');
                      }}
                    >
                      {t('微信')}
                    </Button>
                  </Space>
                </Form>
                {enableStripeTopUp ? (
                    <div>
                      <Form>
                        <Form.Input
                            disabled={!enableStripeTopUp}
                            field={'redemptionCount'}
                            label={t('实付金额：') + ' ' + renderStripeAmount()}
                            placeholder={t('充值数量，最低 ') + stripeMinTopUp + '$'}
                            name='redemptionCount'
                            type={'number'}
                            value={stripeTopUpCount}
                            suffix={'$'}
                            min={stripeMinTopUp}
                            defaultValue={stripeMinTopUp}
                            max={100000}
                            onChange={async (value) => {
                              if (value < 1) {
                                value = 1;
                              }
                              if (value > 100000) {
                                value = 100000;
                              }
                              setStripeTopUpCount(value);
                              await getAmount('stripe', value);
                            }}
                        />
                        <Space>
                          <Button
                              style={{backgroundColor: '#b161fe'}}
                              type={'primary'}
                              disabled={isPaying}
                              theme={'solid'}
                              onClick={async () => {
                                preTopUp('stripe');
                              }}
                          >
                            {isPaying ? '支付中...' : '去支付'}
                          </Button>
                        </Space>
                      </Form>
                    </div>
                ) : (
                    <></>
                )}
              </div>
              {/*<div style={{ display: 'flex', justifyContent: 'right' }}>*/}
              {/*    <Text>*/}
              {/*        <Link onClick={*/}
              {/*            async () => {*/}
              {/*                window.location.href = '/topup/history'*/}
              {/*            }*/}
              {/*        }>充值记录</Link>*/}
              {/*    </Text>*/}
              {/*</div>*/}
            </Card>
          </div>
        </Layout.Content>
      </Layout>
    </div>
  );
};

export default TopUp;
