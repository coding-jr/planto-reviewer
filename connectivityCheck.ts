import fetch from 'node-fetch';

async function main() {
  try {
    const res = await fetch('https://api.github.com/rate_limit', {
      headers: {
        'User-Agent': 'AutoReviewer-Connectivity-Check',
        'Accept': 'application/json'
      },
      // 10s timeout via AbortController
      signal: AbortSignal.timeout(10_000)
    });

    const text = await res.text();
    if (!res.ok) {
      console.error(`Connectivity check failed: HTTP ${res.status} - ${text}`);
      process.exitCode = 1;
      return;
    }

    const data = JSON.parse(text);
    const core = data.resources?.core;
    console.log('Connectivity check result:', {
      core_limit: core?.limit,
      core_remaining: core?.remaining,
      core_reset: core?.reset
    });
  } catch (err) {
    console.error('Connectivity check error:', err);
    process.exitCode = 1;
  }
}

main();

