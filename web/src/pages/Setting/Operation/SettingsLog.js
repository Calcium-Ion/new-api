import React, { useEffect, useState, useRef } from 'react';
import { Button, Col, Form, Row, Spin, DatePicker } from '@douyinfe/semi-ui';
import dayjs from 'dayjs';
import {
  compareObjects,
  API,
  showError,
  showSuccess,
  showWarning,
} from '../../../helpers';

export default function SettingsLog(props) {
  const [loading, setLoading] = useState(false);
  const [loadingCleanHistoryLog, setLoadingCleanHistoryLog] = useState(false);
  const [inputs, setInputs] = useState({
    LogConsumeEnabled: false,
    historyTimestamp: dayjs().subtract(1, 'month').toDate(),
  });
  const refForm = useRef();
  const [inputsRow, setInputsRow] = useState(inputs);

  function onSubmit() {
    const updateArray = compareObjects(inputs, inputsRow).filter(
      (item) => item.key !== 'historyTimestamp',
    );

    if (!updateArray.length) return showWarning('你似乎并没有修改什么');
    const requestQueue = updateArray.map((item) => {
      let value = '';
      if (typeof inputs[item.key] === 'boolean') {
        value = String(inputs[item.key]);
      } else {
        value = inputs[item.key];
      }
      return API.put('/api/option/', {
        key: item.key,
        value,
      });
    });
    setLoading(true);
    Promise.all(requestQueue)
      .then((res) => {
        if (requestQueue.length === 1) {
          if (res.includes(undefined)) return;
        } else if (requestQueue.length > 1) {
          if (res.includes(undefined)) return showError('部分保存失败，请重试');
        }
        showSuccess('保存成功');
        props.refresh();
      })
      .catch(() => {
        showError('保存失败，请重试');
      })
      .finally(() => {
        setLoading(false);
      });
  }
  async function onCleanHistoryLog() {
    try {
      setLoadingCleanHistoryLog(true);
      if (!inputs.historyTimestamp) throw new Error('请选择日志记录时间');
      const res = await API.delete(
        `/api/log/?target_timestamp=${Date.parse(inputs.historyTimestamp) / 1000}`,
      );
      const { success, message, data } = res.data;
      if (success) {
        showSuccess(`${data} 条日志已清理！`);
        return;
      } else {
        throw new Error('日志清理失败：' + message);
      }
    } catch (error) {
      showError(error.message);
    } finally {
      setLoadingCleanHistoryLog(false);
    }
  }

  useEffect(() => {
    const currentInputs = {};
    for (let key in props.options) {
      if (Object.keys(inputs).includes(key)) {
        currentInputs[key] = props.options[key];
      }
    }
    currentInputs['historyTimestamp'] = inputs.historyTimestamp;
    setInputs(Object.assign(inputs, currentInputs));
    setInputsRow(structuredClone(currentInputs));
    refForm.current.setValues(currentInputs);
  }, [props.options]);
  return (
    <>
      <Spin spinning={loading}>
        <Form
          values={inputs}
          getFormApi={(formAPI) => (refForm.current = formAPI)}
          style={{ marginBottom: 15 }}
        >
          <Form.Section text={'日志设置'}>
            <Row gutter={16}>
              <Col span={8}>
                <Form.Switch
                  field={'LogConsumeEnabled'}
                  label={'启用额度消费日志记录'}
                  size='default'
                  checkedText='｜'
                  uncheckedText='〇'
                  onChange={(value) => {
                    setInputs({
                      ...inputs,
                      LogConsumeEnabled: value,
                    });
                  }}
                />
              </Col>
              <Col span={8}>
                <Spin spinning={loadingCleanHistoryLog}>
                  <Form.DatePicker
                    label='日志记录时间'
                    field={'historyTimestamp'}
                    type='dateTime'
                    inputReadOnly={true}
                    onChange={(value) => {
                      setInputs({
                        ...inputs,
                        historyTimestamp: value,
                      });
                    }}
                  />
                  <Button size='default' onClick={onCleanHistoryLog}>
                    清除历史日志
                  </Button>
                </Spin>
              </Col>
            </Row>

            <Row>
              <Button size='default' onClick={onSubmit}>
                保存日志设置
              </Button>
            </Row>
          </Form.Section>
        </Form>
      </Spin>
    </>
  );
}
