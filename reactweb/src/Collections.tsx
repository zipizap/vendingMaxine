import React from 'react';
import { useAuth } from './context/AuthContext';

const Collections = () => {
  // Authentication is now handled by PrivateRoute
  // We can still access auth data if needed
  const { authData } = useAuth();

  return (
    <div>
      <h1>Collections</h1>
      <p>This is the collections webpage.</p>
    </div>
  );
};

export default Collections;
