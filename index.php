<?php
    if (!empty($_GET['data'])) {
        header('Content-Type: application/json;charset=UTF-8');
        echo json_encode([
            0 => ['surname' => 'Иванов', 'name' => 'Иван', 'patronymic' => 'Иванович',],
            1 => ['surname' => 'Петров', 'name' => 'Петр', 'patronymic' => 'Петрович',],
            2 => ['surname' => 'Сидоров', 'name' => 'Василий', 'patronymic' => 'Сергеевич',],
            3 => ['surname' => 'Козлов', 'name' => 'Павел', 'patronymic' => 'Васильевич',],
            4 => ['surname' => 'Кузнецов', 'name' => 'Сергей', 'patronymic' => 'Сергеевич',],
            5 => ['surname' => 'Смирнов', 'name' => 'Виктор', 'patronymic' => 'Андреевич',],
        ]);
    } else {
        phpinfo();
    }