import React, { useState } from 'react';
import axios from 'axios';

const OrderViewer = () => {
    const [orderId, setOrderId] = useState('');
    const [orderData, setOrderData] = useState(null);
    const [error, setError] = useState(null);

    const fetchOrder = async () => {
        try {
            const response = await axios.get(`http://localhost:8080/orders/${orderId}`);
            setOrderData(response.data);
            setError(null);
        } catch (err) {
            setError('Ошибка при получении данных заказа. Проверьте ID и попробуйте снова.');
            setOrderData(null);
        }
    };

    const handleInputChange = (event) => {
        setOrderId(event.target.value);
    };

    const handleSubmit = (event) => {
        event.preventDefault();
        fetchOrder();
    };

    return (
        <div>
            <h1>Просмотр заказа</h1>
            <form onSubmit={handleSubmit}>
                <input 
                    type="text" 
                    value={orderId} 
                    onChange={handleInputChange} 
                    placeholder="Введите ID заказа" 
                    required 
                />
                <button type="submit">Получить заказ</button>
            </form>
            {error && <p style={{ color: 'red' }}>{error}</p>}
            {orderData && (
                <div>
                    <h2>Данные заказа:</h2>
                    <pre>{JSON.stringify(orderData, null, 2)}</pre>
                </div>
            )}
        </div>
    );
};

export default OrderViewer;
