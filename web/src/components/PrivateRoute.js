import { Navigate } from 'react-router-dom';

import { history } from '../helpers';


function PrivateRoute({ children }) {
  if (!localStorage.getItem('user')) {
    return <Navigate to="/login" state={{ from: history.location }} />;
  }
  return children;
}

export { PrivateRoute };