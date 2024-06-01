import React from 'react';
import { Route, Routes } from 'react-router-dom';
import ChannelErrors from './ChannelErrors';

const AdminRoutes = () => {
  return (
    <Routes>
      <Route path="/admin/channel-errors" element={<ChannelErrors />} />
    </Routes>
  );
};

export default AdminRoutes;
