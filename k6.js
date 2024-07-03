import http from 'k6/http';

export const options = {
    // A number specifying the number of VUs to run concurrently.
    vus: 10,
    // A string specifying the total duration of the test run.
    duration: '40s',
}

export default function () {
    const url = 'http://localhost:8080/v1/users';
    const payload = JSON.stringify({
      username: 'user-' + Math.floor(Math.random() * 1000).toString(),
    });
  
    // http.post(url, payload);
    const id = http.post(url, payload).json().id;

    http.get(`${url}/${id}`);

    for (let i = 0; i < 40; i++) {
        http.get(url);
    }
}
