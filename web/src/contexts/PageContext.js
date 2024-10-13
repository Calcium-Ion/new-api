import React, { createContext, useState, useContext } from 'react';

const PageContext = createContext();

export const PageProvider = ({ children }) => {
  const [isChat, setIsChat] = useState(false);

  return (
    <PageContext.Provider value={{ isChat, setIsChat }}>
      {children}
    </PageContext.Provider>
  );
};

export const usePageContext = () => useContext(PageContext);
