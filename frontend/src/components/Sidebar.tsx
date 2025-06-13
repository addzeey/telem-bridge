import { Link } from '@tanstack/react-router';
import '@styles/sidebar.css'; // Import your CSS for styling
import { useTheme } from './ThemeProvider';
import { useEffect, useState } from 'react';

// Define the navigation structure as a tree
const navItems: NavItem[] = [
	{
		label: 'Home',
		to: '/',
	},
	{
		label: 'Settings',
		to: '/settings',
		children: [
			{
				label: 'Packet Forwarding',
				to: '/settings/telemetry',
			},
      {
        label: 'OSC Addresses',
        to: '/settings/osc',
      },
		],
	},
    {
        label: 'Telemetry',
        to: '/telemetry/live',
        children: [
            {
                label: 'Live Data',
                to: '/telemetry/live',
            },
        ],
    },
	// Add more top-level or nested items as needed
];

// Add a type for nav items
interface NavItem {
	label: string;
	to: string;
	children?: NavItem[];
}

export function Sidebar() {
	const { theme, toggleTheme } = useTheme();
	const [version, setVersion] = useState<string>('');

	useEffect(() => {
		fetch('/api/version')
			.then(r => r.json())
			.then(data => setVersion(data.version || ''));
	}, []);

	// Recursive render function for nav items
	function renderNav(items: NavItem[], depth = 0) {
		return items.map(item => (
			<div key={item.to} className={`sidebar-nav-item d-flex flex-column gap-2 depth-${depth}`}>
				<Link
					to={item.to}
					activeProps={{ className: 'btn-primary active-nav' }}
					inactiveProps={{ className: 'btn-secondary' }}
					className="btn text-white w-100 text-start"
				>
					{item.label}
				</Link>
				{item.children && renderNav(item.children, depth + 1)}
			</div>
		));
	}

	return (
		<aside className="sidebar-fixed bg-dark text-white d-flex flex-column p-3">
			<h2 className="fs-5 fw-bold mb-4">F1 Telemetry</h2>
			<nav className="nav flex-column gap-2">{renderNav(navItems)}</nav>
			<div className="mt-auto pt-3">
				<button
					onClick={toggleTheme}
					className="btn btn-outline-light w-100"
					type="button"
					title="Toggle dark/light mode"
				>
					{theme === 'dark' ? '🌙 Dark Mode' : '☀️ Light Mode'}
				</button>
				{version && (
					<div className="text-center mt-3 small text-secondary">
						Version {version}
					</div>
				)}
			</div>
		</aside>
	);
}
