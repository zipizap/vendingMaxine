import React, { useEffect, useState } from 'react';
import './App.css';

interface AuthData {
  authenticated: boolean;
  claims: {
    [key: string]: any;
  };
}

const Profile = () => {
  const [isAuthenticated, setIsAuthenticated] = useState<boolean | null>(null);
  const [authData, setAuthData] = useState<AuthData | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const checkAuth = async () => {
      try {
        const response = await fetch('/api/public/check_auth');
        const data = await response.json();
        
        setIsAuthenticated(data.authenticated);
        setAuthData(data);
        
        if (!data.authenticated) {
          const currentUrl = encodeURIComponent(window.location.href);
          window.location.href = `/login?redirect=${currentUrl}`;
          return;
        }
        
        setIsLoading(false);
      } catch (error) {
        console.error('Profile: Authentication check failed with error:', error);
        const currentUrl = encodeURIComponent(window.location.href);
        window.location.href = `/login?redirect=${currentUrl}`;
      }
    };

    checkAuth();
  }, []);

  if (isLoading || isAuthenticated === null) {
    return <div className="profile-loading">Loading...</div>;
  }

  // Format dates for better readability
  const formatDate = (timestamp: number): string => {
    return new Date(timestamp * 1000).toLocaleString();
  };

  // Function to render the auth data in a well-formatted way
  const renderAuthData = () => {
    if (!authData || !authData.claims) return null;

    return Object.entries(authData.claims).map(([key, value]) => {
      // Format the value based on its type
      let formattedValue: React.ReactNode;

      if (key === 'exp' || key === 'iat') {
        formattedValue = formatDate(value as number);
      } else if (Array.isArray(value)) {
        formattedValue = (
          <ul className="profile-list">
            {value.map((item, i) => (
              <li key={i}>{item}</li>
            ))}
          </ul>
        );
      } else if (typeof value === 'boolean') {
        formattedValue = value ? 'Yes' : 'No';
      } else {
        formattedValue = String(value);
      }

      return (
        <div className="profile-item" key={key}>
          <div className="profile-key">{formatKeyName(key)}:</div>
          <div className="profile-value">{formattedValue}</div>
        </div>
      );
    });
  };

  // Function to format key names for better readability
  const formatKeyName = (key: string): string => {
    // Handle special cases
    switch(key) {
      case 'iat': return 'Issued At';
      case 'exp': return 'Expires At';
      case 'sub': return 'Subject';
      case 'iss': return 'Issuer';
      case 'aud': return 'Audience';
      default:
        // Convert camelCase or snake_case to Title Case With Spaces
        return key
          .replace(/_/g, ' ')
          .replace(/([A-Z])/g, ' $1')
          .replace(/^./, str => str.toUpperCase());
    }
  };

  return (
    <div className="profile-container">
      <h1>User Profile</h1>
      
      <div className="profile-section">
        <h2>Authentication Information</h2>
        <div className="profile-data">
          <div className="profile-item">
            <div className="profile-key">Authentication Status:</div>
            <div className="profile-value authenticated">
              {isAuthenticated ? 'Authenticated' : 'Not Authenticated'}
            </div>
          </div>
          
          {renderAuthData()}
        </div>
      </div>
    </div>
  );
};

export default Profile;
