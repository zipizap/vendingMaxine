import React, { useEffect } from 'react';
import { Navigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';

interface PrivateRouteProps {
  children: React.ReactNode;
}

const PrivateRoute: React.FC<PrivateRouteProps> = ({ children }) => {
  const { isAuthenticated, isLoading, handleLogin } = useAuth();
  
  useEffect(() => {
    if (!isLoading && !isAuthenticated) {
      handleLogin();
    }
  }, [isAuthenticated, isLoading, handleLogin]);

  if (isLoading) {
    return <div>Loading...</div>;
  }
  
  return isAuthenticated ? <>{children}</> : null;
};

export default PrivateRoute;
