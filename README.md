Note: This Project is for Learning Purposes
Please be aware that the project is designed solely for learning purposes. It is not intended to be used in production environments 


# **Sql-mi**

sql-mi is a tool designed to convert a schema for defining database tables into SQL code. This tool currently supports generating SQL code for SQLite databases. It enables you to define tables and their attributes in a simplified syntax and convert them into corresponding SQL statements.

## Installation

To get started, download the specific version of sql-mi that matches your system from the [Releases](https://github.com/Blackarrow299/sql-mi/releases) section. After downloading, follow these steps to set up the tool on your system:

1. Unzip the downloaded file.

2. Open a terminal window and navigate to the directory containing the unzipped files.

3. Make the binary executable:

   ```bash
   chmod +x sql-mi
   ```

4. Optionally, move the binary to a directory in your system's `PATH` to make it accessible from anywhere.

## Usage

To generate SQL code from your schema, follow this pattern:

```bash
./sql-mi -o output.sql input
```

Where:
- `output.sql`: The name of the output SQL file where the generated SQL code will be saved.
- `input`: The name of the input file containing your schema.

### Example

Suppose you have a file named `schema.txt` containing the following syntax:

```plaintext
table users
	id int @id @auto_increment
	created_at datetime @default(`NOW()`)
end

table contacts
	name `varchar(255)` @default("text") @nullable
	user_id int @reference("users", "id") @onDelete("RESTRICT") @onUpdate("CASCADE")
	created_at datetime @default(`NOW()`)
end
```

You can generate the corresponding SQL code by running:

```bash
./sql-mi -o output.sql schema.txt
```

This will create an `output.sql` file containing the generated SQL statements.

## Supported Attributes

Sql-mi supports the following attributes for table columns:

- `@default`: Set the default value for the column. You can enter raw SQL like @default(\`NOW()\`) to use SQL functions.
- `@id`: Mark the column as the primary key.
- `@auto_increment`: Enable auto-increment for integer columns.
- `@nullable`: Allow null values for the column.
- `@reference`: Define a foreign key reference to another table.
- `@onDelete`: Specify the behavior on delete (e.g., "RESTRICT", "CASCADE").
- `@onUpdate`: Specify the behavior on update (e.g., "RESTRICT", "CASCADE").

## Supported Data Types

Sql-mi supports the following data types:

- `int`
- `string`
- `datetime`
- `bool`
- `float`
- `blob`

Additionally, you can enter raw SQL data types like \`varchar(255)\`.

## Database Providers

The project currently supports SQLite databases. You can specify the database provider by implementing the following syntax in your input file:

```
set provider sqlite
```

## Contributions

Contributions to this project are welcome! If you have ideas for improvements or new features, feel free to open an issue or submit a pull request.

---
