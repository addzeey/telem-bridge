import { createFileRoute } from '@tanstack/react-router';
import { useEffect, useRef, useState } from 'react';
import { FontAwesomeIcon } from '@fortawesome/react-fontawesome'
import { faCopy } from '@fortawesome/free-solid-svg-icons'
import 'bootstrap/dist/css/bootstrap.min.css';

interface OSCAddressEntry {
    address: string;
    type: string;
    enabled: boolean;
    allowZero?: boolean; // Add allowZero property
}

// Group addresses by prefix (e.g. /car/, /motion/)
function groupByPrefix(addresses: Record<string, OSCAddressEntry>) {
    const groups: Record<string, { label: string; keys: string[]; }> = {};
    Object.entries(addresses).forEach(([key, entry]) => {
        const match = entry.address.match(/^\/(\w+)/);
        const group = match ? match[1] : 'Other';
        if (!groups[group]) {
            groups[group] = { label: group.charAt(0).toUpperCase() + group.slice(1), keys: [] };
        }
        groups[group].keys.push(key);
    });
    return groups;
}

export const Route = createFileRoute('/settings/osc')({
    component: OSCSettingsPage,
});

function OSCSettingsPage() {
    const [oscAddresses, setOscAddresses] = useState<Record<string, OSCAddressEntry> | null>(null);
    const [saveStatus, setSaveStatus] = useState<null | 'saving' | 'saved'>(null);
    const [copiedKey, setCopiedKey] = useState<string | null>(null);
    const copyTimeout = useRef<number | null>(null);
    const saveTimeout = useRef<number | null>(null);
    const toastRef = useRef<HTMLDivElement | null>(null);

    useEffect(() => {
        fetch('/api/osc-addresses')
            .then(r => r.json())
            .then(setOscAddresses);
    }, []);

    useEffect(() => {
        if (saveStatus && toastRef.current) {
            // Show toast when saveStatus changes
            const toastEl = toastRef.current;
            toastEl.classList.add('show');
            toastEl.classList.remove('hide');
            // Hide after 1.5s if saved
            if (saveStatus === 'saved') {
                setTimeout(() => {
                    toastEl.classList.remove('show');
                    toastEl.classList.add('hide');
                }, 1500);
            }
        }
    }, [saveStatus]);

    function saveAddresses(addresses: Record<string, OSCAddressEntry>) {
        setSaveStatus('saving');
        fetch('/api/osc-addresses', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(addresses),
        }).then(() => {
            setSaveStatus('saved');
            if (saveTimeout.current) clearTimeout(saveTimeout.current);
            saveTimeout.current = setTimeout(() => setSaveStatus(null), 1500);
        });
    }

    function handleToggle(key: string) {
        if (!oscAddresses) return;
        const updated = { ...oscAddresses, [key]: { ...oscAddresses[key], enabled: !oscAddresses[key].enabled } };
        setOscAddresses(updated);
        saveAddresses(updated);
    }

    function handleParentToggle(group: string, enable: boolean) {
        if (!oscAddresses) return;
        const grouped = groupByPrefix(oscAddresses);
        const keys = grouped[group]?.keys || [];
        const updated = { ...oscAddresses };
        keys.forEach(k => {
            updated[k] = { ...updated[k], enabled: enable };
        });
        setOscAddresses(updated);
        saveAddresses(updated);
    }

    function handleCopy(address: string, key: string) {
        navigator.clipboard.writeText(address);
        setCopiedKey(key);
        if (copyTimeout.current) clearTimeout(copyTimeout.current);
        copyTimeout.current = setTimeout(() => setCopiedKey(null), 1200);
    }

    function handleAllowZeroToggle(key: string) {
        if (!oscAddresses) return;
        const updated = {
            ...oscAddresses,
            [key]: { ...oscAddresses[key], allowZero: !oscAddresses[key].allowZero }
        };
        setOscAddresses(updated);
        saveAddresses(updated);
    }

    const groups = oscAddresses ? groupByPrefix(oscAddresses) : {};

    return (
        <div className="container py-4 position-relative">
            {/* Spacer for alert */}
            <div style={{ height: '44px' }}></div>
            {saveStatus === 'saving' && (
                <div
                    className="alert alert-info py-2 position-fixed w-100 top-0 start-0 text-center"
                    style={{ zIndex: 2000, left: 0, right: 0, width: '100vw' }}
                >
                    Saving...
                </div>
            )}
            {saveStatus === 'saved' && (
                <div
                    className="alert alert-success py-2 position-fixed w-100 top-0 start-0 text-center"
                    style={{ zIndex: 2000, left: 0, right: 0, width: '100vw' }}
                >
                    Saved!
                </div>
            )}
            <h1 className="h3 mb-4">OSC Address Settings</h1>
            {!oscAddresses ? (
                <div className="alert alert-info">Loading...</div>
            ) : (
                <div className="mb-4">
                    {Object.entries(groups).map(([group, { label, keys }]) => {
                        const allEnabled = keys.every(k => oscAddresses[k].enabled);
                        const someEnabled = keys.some(k => oscAddresses[k].enabled);
                        return (
                            <div key={group} className="mb-3 border rounded p-2 bg-secondary-subtle">
                                <h4>{label}</h4>
                                <div className="form-check mb-2">
                                    <input
                                        className="form-check-input"
                                        type="checkbox"
                                        id={`parent-${group}`}
                                        checked={allEnabled}
                                        ref={el => { if (el) el.indeterminate = !allEnabled && someEnabled; }}
                                        onChange={e => handleParentToggle(group, e.target.checked)}
                                    />
                                    <label className="form-check-label fw-bold" htmlFor={`parent-${group}`}>
                                        {label} (All)
                                    </label>
                                </div>
                                <div className="row g-2">
                                    {keys.map(key => (
                                        <div className="col-12 col-lg-6" key={key}>
                                            <div className="form-check d-flex align-items-center mb-1">
                                                <input
                                                    className="form-check-input me-2"
                                                    type="checkbox"
                                                    id={`osc-${key}`}
                                                    checked={oscAddresses[key].enabled}
                                                    onChange={() => handleToggle(key)}
                                                />
                                                <label className="form-check-label w-75 d-flex align-items-center" htmlFor={`osc-${key}`}>
                                                    <span className="fw-semibold">{oscAddresses[key].address}</span>
                                                    <span className="badge bg-secondary ms-2">{oscAddresses[key].type}</span>
                                                    <button
                                                        type="button"
                                                        className="btn btn-sm btn-outline-secondary ms-2"
                                                        title="Copy OSC address"
                                                        style={{padding: '0.15rem 0.5rem', fontSize: '0.9em'}} 
                                                        onClick={e => { e.preventDefault(); handleCopy(oscAddresses[key].address, key); }}
                                                    >
                                                        <FontAwesomeIcon icon={faCopy} />
                                                    </button>
                                                    {copiedKey === key && <span className="ms-2 text-success small">Copied!</span>}
                                                </label>
                                                <div className="allow-zero">
                                                                                                    {/* Allow Zero Checkbox */}
                                                <input
                                                    className="form-check-input ms-2"
                                                    type="checkbox"
                                                    id={`allowzero-${key}`}
                                                    checked={!!oscAddresses[key].allowZero}
                                                    onChange={() => handleAllowZeroToggle(key)}
                                                    title="Allow sending zero values for this OSC address"
                                                />
                                                <label className="form-check-label ms-1" htmlFor={`allowzero-${key}`}>
                                                    Allow Zero
                                                </label>
                                                </div>
                                            </div>
                                        </div>
                                    ))}
                                </div>
                            </div>
                        );
                    })}
                </div>
            )}
        </div>
    );
}
