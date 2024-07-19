import { getUserIdFromLocalStorage, showError } from './utils';
import axios from 'axios';

export const API = axios.create({
  baseURL: import.meta.env.VITE_REACT_APP_SERVER_URL
    ? import.meta.env.VITE_REACT_APP_SERVER_URL
    : '',
  headers: {
    'New-API-User': getUserIdFromLocalStorage()
  }
});

API.interceptors.response.use(
  (response) => response,
  (error) => {
    showError(error);
  },
);
