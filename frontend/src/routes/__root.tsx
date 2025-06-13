import { createRootRoute, Outlet, useRouterState } from '@tanstack/react-router';
import { Sidebar } from '../components/Sidebar';
import '@styles/App.css';

export const Route = createRootRoute({
  component: RootLayout,
});

function RootLayout() {
  const { location } = useRouterState();
  const hideSidebar = location.pathname.startsWith('/overlay');
  return (
    <div className="d-flex vh-100">
      {!hideSidebar && <Sidebar />}
      <main className="flex-grow-1 overflow-auto" style={{ background: 'var(--bs-body-bg)', color: 'var(--bs-body-color)' }}>
        <Outlet />
      </main>
    </div>
  );
}