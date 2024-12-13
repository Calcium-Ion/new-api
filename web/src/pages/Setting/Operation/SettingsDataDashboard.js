import React, { useEffect, useState, useRef } from 'react';
import { Button, Col, Form, Row, Spin } from '@douyinfe/semi-ui';
import {
  compareObjects,
  API,
  showError,
  showSuccess,
  showWarning,
} from '../../../helpers';
import { useTranslation } from 'react-i18next';

export default function DataDashboard(props) {
  const { t } = useTranslation();
  
  const optionsDataExportDefaultTime = [
    { key: 'hour', label: t('小时'), value: 'hour' },
    { key: 'day', label: t('天'), value: 'day' },
    { key: 'week', label: t('周'), value: 'week' },
  ];
  const [loading, setLoading] = useState(false);
  const [inputs, setInputs] = useState({
    DataExportEnabled: false,
    DataExportInterval: '',
    DataExportDefaultTime: '',
  });
  const refForm = useRef();
  const [inputsRow, setInputsRow] = useState(inputs);

  function onSubmit() {
    const updateArray = compareObjects(inputs, inputsRow);
    if (!updateArray.length) return showWarning(t('你似乎并没有修改什么'));
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
          if (res.includes(undefined)) return showError(t('部分保存失败，请重试'));
        }
        showSuccess(t('保存成功'));
        props.refresh();
      })
      .catch(() => {
        showError(t('保存失败，请重试'));
      })
      .finally(() => {
        setLoading(false);
      });
  }

  useEffect(() => {
    const currentInputs = {};
    for (let key in props.options) {
      if (Object.keys(inputs).includes(key)) {
        currentInputs[key] = props.options[key];
      }
    }
    setInputs(currentInputs);
    setInputsRow(structuredClone(currentInputs));
    refForm.current.setValues(currentInputs);
    localStorage.setItem(
      'data_export_default_time',
      String(inputs.DataExportDefaultTime),
    );
  }, [props.options]);

  return (
    <>
      <Spin spinning={loading}>
        <Form
          values={inputs}
          getFormApi={(formAPI) => (refForm.current = formAPI)}
          style={{ marginBottom: 15 }}
        >
          <Form.Section text={t('数据看板设置')}>
            <Row gutter={16}>
              <Col span={8}>
                <Form.Switch
                  field={'DataExportEnabled'}
                  label={t('启用数据看板（实验性）')}
                  size='default'
                  checkedText='｜'
                  uncheckedText='〇'
                  onChange={(value) => {
                    setInputs({
                      ...inputs,
                      DataExportEnabled: value,
                    });
                  }}
                />
              </Col>
            </Row>
            <Row>
              <Col span={8}>
                <Form.InputNumber
                  label={t('数据看板更新间隔')}
                  step={1}
                  min={1}
                  suffix={t('分钟')}
                  extraText={t('设置过短会影响数据库性能')}
                  placeholder={t('数据看板更新间隔')}
                  field={'DataExportInterval'}
                  onChange={(value) =>
                    setInputs({
                      ...inputs,
                      DataExportInterval: String(value),
                    })
                  }
                />
              </Col>
              <Col span={8}>
                <Form.Select
                  label={t('数据看板默认时间粒度')}
                  optionList={optionsDataExportDefaultTime}
                  field={'DataExportDefaultTime'}
                  extraText={t('仅修改展示粒度，统计精确到小时')}
                  placeholder={t('数据看板默认时间粒度')}
                  style={{ width: 180 }}
                  onChange={(value) =>
                    setInputs({
                      ...inputs,
                      DataExportDefaultTime: String(value),
                    })
                  }
                />
              </Col>
            </Row>
            <Row>
              <Button size='default' onClick={onSubmit}>
                {t('保存数据看板设置')}
              </Button>
            </Row>
          </Form.Section>
        </Form>
      </Spin>
    </>
  );
}
