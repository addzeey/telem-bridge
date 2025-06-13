import React, { createContext, useContext, useEffect, useState } from 'react';

// Helper to get/set cookie
function getCookie(name: string): string | null {
  const match = document.cookie.match(new RegExp('(^| )' + name + '=([^;]+)'));
  return match ? decodeURIComponent(match[2]) : null;
}
function setCookie(name: string, value: string, days = 365) {
  const expires = new Date(Date.now() + days * 864e5).toUTCString();
  document.cookie = `${name}=${encodeURIComponent(value)}; expires=${expires}; path=/`;
}

// Theme context
const ThemeContext = createContext<{
  theme: 'dark' | 'light',
  toggleTheme: () => void
}>({ theme: 'dark', toggleTheme: () => {} });

export const useTheme = () => useContext(ThemeContext);

export const ThemeProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [theme, setTheme] = useState<'dark' | 'light'>(() => {
    const cookie = getCookie('theme');
    return cookie === 'light' ? 'light' : 'dark'; // default to dark
  });

  useEffect(() => {
    document.documentElement.classList.remove('light', 'dark');
    document.body.classList.remove('light', 'dark');
    document.documentElement.classList.add(theme);
    document.body.classList.add(theme);
    document.documentElement.setAttribute('data-bs-theme', theme); // Set Bootstrap theme attribute
    setCookie('theme', theme);
  }, [theme]);

  const toggleTheme = () => setTheme(t => (t === 'dark' ? 'light' : 'dark'));

  return (
    <ThemeContext.Provider value={{ theme, toggleTheme }}>
      {children}
    </ThemeContext.Provider>
  );
};
