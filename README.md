Для создания базы данных через миграции ввести команду migrate `-path db/migrations -database "postgres://postgres:password@localhost:5432/users?sslmode=disable" up` находясь в папке `reward-service/cmd/api`  
Для запуска в Docker'e необходимо ввести команду в консоль `make up_build` внутри папки project  
Наличие требования для access token'a:  
![access_through_access_token](https://github.com/user-attachments/assets/cfeac453-6c2b-4a62-9306-900c4250b0d8)  
  
Создание нового user'a с указанными данными, пароль хранится в PostgresSQL в хэшированном виде:  
  
![create_user](https://github.com/user-attachments/assets/b6728319-b417-42e9-882d-ff6883469da9)  

Аутентификация только что созданным пользователем, после чего ему автоматически генерируется access token:  
  
![auth_user](https://github.com/user-attachments/assets/fd2eb4ee-6d5f-476b-bea2-19792cd93906)  
  
Статус пользователя, определяется путём id в строке запроса:  
  
![user_status](https://github.com/user-attachments/assets/1864f19d-c19c-428a-8010-e50082bd25c5)  

Таблица с пользователями, отсортированы по убыванию от пользователя с наибольшим кол-вом очков:  
  
![user_leaderboard](https://github.com/user-attachments/assets/7bf4412f-fbc2-43e7-aa07-8d4ef4b6a35d)  

Пример выполнения заданий по регистрации в Телеграм и Твиттер, а также их результаты:  
  
![X_Sign_8](https://github.com/user-attachments/assets/65af356b-f431-4937-94b1-a91cf6749dba)  
![telegram_sign_8](https://github.com/user-attachments/assets/27ff75e4-7c16-453a-bcdb-e99ff1df06db)  
![result_after_signing_8](https://github.com/user-attachments/assets/7a093e1c-fa31-4fc9-86a3-5ee9c7bde4bf)  
  
Пример выполнения какого-либо кастомного задания, поинты можно задать в теле запроса и его результат:  
  
![custom_task_3](https://github.com/user-attachments/assets/a09120e5-ffad-470f-94d9-34fc86d90f4e)  
![result_after_custom_task](https://github.com/user-attachments/assets/fdb0e399-0aaa-4a88-b985-b59b99d857ac)  

Пример использования ваучера. При использовании его, в запросе передаётся id пользователя, кто им воспользовался, а в теле сам ваучер. 
Пользователю по id начисляется 25 очков, а пользователю, который привёл нового пользователя по ваучеру, начисляется 100 очков:  

  ![referrer_success](https://github.com/user-attachments/assets/6aac0514-1577-4c64-bdd1-81d19eef179a)  
  ![referrer_success_show](https://github.com/user-attachments/assets/20002384-45b3-491b-8508-133bde1e6cbe)  

Также пример некоторых запросов, когда access token не был получен:  

![изображение](https://github.com/user-attachments/assets/8a8c33da-a845-45a1-8ca0-4948dd5f2d83)  
![изображение](https://github.com/user-attachments/assets/a9ff9e7d-a4e0-4c2c-a09f-17f94dd84aa3)  

