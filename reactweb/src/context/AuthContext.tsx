import React, { createContext, useState, useEffect, useContext, ReactNode } from 'react';

interface AuthData {
  authenticated: boolean;
  claims?: {
    name?: string;
    email?: string;
    [key: string]: any;
  };
}

interface AuthContextType {
  isAuthenticated: boolean;
  isLoading: boolean;
  userName: string;
  userEmail: string;
  authData: AuthData | null;
  checkAuth: () => Promise<boolean>;
  handleLogin: () => void;
  handleLogout: () => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const AuthProvider: React.FC<{ children: ReactNode }> = ({ children }) => {
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [isLoading, setIsLoading] = useState(true);
  const [authData, setAuthData] = useState<AuthData | null>(null);
  const [userName, setUserName] = useState('');
  const [userEmail, setUserEmail] = useState('');

  const checkAuth = async (): Promise<boolean> => {
    try {
      const response = await fetch('/api/public/check_auth');
      const data = await response.json();
      
      setIsAuthenticated(data.authenticated);
      setAuthData(data);
      
      if (data.authenticated && data.claims) {
        if (data.claims.name) {
          setUserName(data.claims.name);
        }
        if (data.claims.email) {
          setUserEmail(data.claims.email);
        }
      }
      
      setIsLoading(false);
      return data.authenticated;
    } catch (error) {
      console.error('Error checking authentication:', error);
      setIsLoading(false);
      setIsAuthenticated(false);
      return false;
    }
  };

  const handleLogin = () => {
    const currentUrl = encodeURIComponent(window.location.href);
    window.location.href = `/login?redirect=${currentUrl}`;
  };

  const handleLogout = () => {
    window.location.href = '/logout';
  };

  useEffect(() => {
    checkAuth();
  }, []);

  return (
    <AuthContext.Provider value={{ 
      isAuthenticated, 
      isLoading, 
      userName, 
      userEmail, 
      authData,
      checkAuth,
      handleLogin,
      handleLogout
    }}>
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
};
