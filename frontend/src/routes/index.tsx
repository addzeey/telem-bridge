import { createFileRoute } from '@tanstack/react-router';
import { ServiceRestartButton } from '../components/ServiceRestartButton';

export const Route = createFileRoute('/')({
    component: HomePage,
});

function HomePage() {
    return (
        <div className="container py-4">
            <h1 className="display-5 fw-bold mb-4">F1 Telemetry Bridge Dashboard</h1>
            <p className="lead">This is the home page. Use the sidebar to navigate.</p>
            <div className="mb-4 d-flex flex-row gap-2">
                <ServiceRestartButton service="udp" />
                <ServiceRestartButton service="osc" />
                <ServiceRestartButton service="all" label="Restart All Services" />
            </div>
        </div>
    );
}