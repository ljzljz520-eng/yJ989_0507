const status = document.querySelector("#status");

async function showServiceStatus() {
  try {
    const response = await fetch("/", { headers: { Accept: "application/json" } });
    const payload = await response.json();
    status.textContent = `${payload.service}: ${payload.status}`;
  } catch {
    status.textContent = "Go service is not connected";
  }
}

showServiceStatus();
