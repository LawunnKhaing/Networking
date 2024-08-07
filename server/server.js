import express from "express";
import { InfluxDB, Point } from '@influxdata/influxdb-client'
import dotenv from "dotenv";

dotenv.config();
// Create an Express application
const app = express();

const PORT = 6969;

// InfluxDB configuration
const url = process.env.INFLUX_URL;
const token = process.env.INFLUX_TOKEN;
const org = process.env.INFLUX_ORG;
const bucket = process.env.INFLUX_BUCKET;

const influxDB = new InfluxDB({ url, token });

const writeApi = influxDB.getWriteApi(org, bucket);


// Route to handle incoming data
app.post('/data', (req, res) => {
    const { location, temperature, humidity } = req.body;

    if (!location || typeof temperature !== 'number' || typeof humidity !== 'number') {
        return res.status(400).json({ error: 'Invalid input data' });
    }

    const point = new Point('weather')
        .tag('location', location)
        .floatField('temperature', temperature)
        .floatField('humidity', humidity);

    writeApi.writePoint(point);
    writeApi
        .flush()
        .then(() => {
            console.log('Data written to InfluxDB');
            res.status(200).json({ message: 'Data written successfully' });
        })
        .catch(error => {
            console.error('Error writing data to InfluxDB:', error);
            res.status(500).json({ error: 'Error writing data to InfluxDB' });
        });
});

// Start the Express server
app.listen(PORT, () => {
    console.log(`Server is running on port ${PORT}`);
});
