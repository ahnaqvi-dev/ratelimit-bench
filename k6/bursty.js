import http from 'k6/http';

// quiet baseline with a burst every 20 seconds, 200 simulated clients
export const options = {
  scenarios: {
    bursty: {
      executor: 'ramping-arrival-rate',
      startRate: 2000,
      timeUnit: '1s',
      preAllocatedVUs: 400,
      maxVUs: 2000,
      stages: [
        { target: 2000, duration: '15s' },
        { target: 15000, duration: '2s' },
        { target: 15000, duration: '3s' },
        { target: 2000, duration: '2s' },
        { target: 2000, duration: '13s' },
        { target: 15000, duration: '2s' },
        { target: 15000, duration: '3s' },
        { target: 2000, duration: '2s' },
      ],
    },
  },
};

export default function () {
  const client = `c${Math.floor(Math.random() * 200)}`;
  http.get('http://localhost:8080/', { headers: { 'X-Client': client } });
}
