// contexts/User/index.jsx

import React from 'react';
import { initialState, reducer } from './reducer';

export const StatusContext = React.createContext({
  state: initialState,
  dispatch: () => null,
});

export const StatusProvider = ({ children }) => {
  const [state, dispatch] = React.useReducer(reducer, initialState);

  return (
    <StatusContext.Provider value={[state, dispatch]}>
      {children}
    </StatusContext.Provider>
  );
};
