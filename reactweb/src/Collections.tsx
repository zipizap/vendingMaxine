import React, { useEffect, useState } from 'react';

const Collections = () => {
  const [isAuthenticated, setIsAuthenticated] = useState<boolean | null>(null);

  useEffect(() => {
    console.log('Collections: Starting authentication check');
    
    const checkAuth = async () => {
      try {
        console.log('Collections: Fetching authentication status');
        const response = await fetch('/api/public/check_auth');
        console.log('Collections: API response status:', response.status);
        
        const data = await response.json();
        console.log('Collections: Auth data received:', data);
        
        if (!data.authenticated) {
          console.log('Collections: User not authenticated, redirecting to login');
          const currentUrl = encodeURIComponent(window.location.href);
          window.location.href = `/login?redirect=${currentUrl}`;
          return;
        }
        
        console.log('Collections: User is authenticated, setting state');
        setIsAuthenticated(true);
      } catch (error) {
        console.error('Collections: Authentication check failed with error:', error);
        console.log('Collections: Redirecting to login due to error');
        const currentUrl = encodeURIComponent(window.location.href);
        window.location.href = `/login?redirect=${currentUrl}`;
      }
    };

    checkAuth();
    
    return () => {
      console.log('Collections: Cleanup of authentication check effect');
    };
  }, []);

  // Show loading state while checking authentication
  if (isAuthenticated === null) {
    console.log('Collections: Rendering loading state');
    return <div>Loading...</div>;
  }

  console.log('Collections: Rendering authenticated content');
  // Only render content if authenticated
  return (
    <div>
      <h1>Collections</h1>
      <p>This is the collections webpage.</p>
    </div>
  );
};

export default Collections;
