import { useEffect, useState } from "react";

export type TelemetryData = {
	Speed: number;
	Throttle: number;
	Steer: number;
	Brake: number;
	Clutch: number;
	Gear: number;
	RPM: number;
};

export function useTelemetry() {
	const [data, setData] = useState<TelemetryData | null>(null);

	useEffect(() => {
		const ws = new WebSocket("ws://localhost:1337/ws");

		ws.onmessage = (event) => {
			try {
				const parsed: TelemetryData = JSON.parse(event.data);
				setData(parsed);
			} catch (err) {
				console.error("WebSocket JSON error", err);
			}
		};

		ws.onclose = () => {
			console.warn("WebSocket connection closed");
		};

		return () => {
			ws.close();
		};
	}, []);

	return data;
}
