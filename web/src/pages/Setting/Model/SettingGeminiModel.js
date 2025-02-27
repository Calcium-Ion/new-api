import React, { useEffect, useState, useRef } from 'react';
import { Button, Col, Form, Row, Spin } from '@douyinfe/semi-ui';
import {
  compareObjects,
  API,
  showError,
  showSuccess,
  showWarning, verifyJSON
} from '../../../helpers';
import { useTranslation } from 'react-i18next';

const GEMINI_SETTING_EXAMPLE = {
  'default': 'OFF',
  'HARM_CATEGORY_CIVIC_INTEGRITY': 'BLOCK_NONE',
};

const GEMINI_VERSION_EXAMPLE = {
  'default': 'v1beta',
};


export default function SettingGeminiModel(props) {
  const { t } = useTranslation();

  const [loading, setLoading] = useState(false);
  const [inputs, setInputs] = useState({
    'gemini.safety_settings': '',
    'gemini.version_settings': '',
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
  }, [props.options]);

  return (
    <>
      <Spin spinning={loading}>
        <Form
          values={inputs}
          getFormApi={(formAPI) => (refForm.current = formAPI)}
          style={{ marginBottom: 15 }}
        >
          <Form.Section text={t('Gemini设置')}>
            <Row>
              <Col span={16}>
                <Form.TextArea
                  label={t('Gemini安全设置')}
                  placeholder={t('为一个 JSON 文本，例如：') + '\n' + JSON.stringify(GEMINI_SETTING_EXAMPLE, null, 2)}
                  field={'gemini.safety_settings'}
                  extraText={t('default为默认设置，可单独设置每个分类的安全等级')}
                  autosize={{ minRows: 6, maxRows: 12 }}
                  trigger='blur'
                  stopValidateWithError
                  rules={[
                    {
                      validator: (rule, value) => verifyJSON(value),
                      message: t('不是合法的 JSON 字符串')
                    }
                  ]}
                  onChange={(value) => setInputs({ ...inputs, 'gemini.safety_settings': value })}
                />
              </Col>
            </Row>
            <Row>
              <Col span={16}>
                <Form.TextArea
                  label={t('Gemini版本设置')}
                  placeholder={t('为一个 JSON 文本，例如：') + '\n' + JSON.stringify(GEMINI_VERSION_EXAMPLE, null, 2)}
                  field={'gemini.version_settings'}
                  extraText={t('default为默认设置，可单独设置每个模型的版本')}
                  autosize={{ minRows: 6, maxRows: 12 }}
                  trigger='blur'
                  stopValidateWithError
                  rules={[
                    {
                      validator: (rule, value) => verifyJSON(value),
                      message: t('不是合法的 JSON 字符串')
                    }
                  ]}
                  onChange={(value) => setInputs({ ...inputs, 'gemini.version_settings': value })}
                />
              </Col>
            </Row>

            <Row>
              <Button size='default' onClick={onSubmit}>
                {t('保存')}
              </Button>
            </Row>
          </Form.Section>
        </Form>
      </Spin>
    </>
  );
}
