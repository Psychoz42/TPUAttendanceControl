package com.example.tputracking.presentation.login.components

import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.OutlinedTextFieldDefaults
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.input.VisualTransformation
import androidx.compose.ui.unit.dp
import com.example.tputracking.ui.theme.BorderLight
import com.example.tputracking.ui.theme.GreenPrimary
import com.example.tputracking.ui.theme.TextSecondary

@Composable
fun LoginTextField(
    modifier: Modifier = Modifier,
    value: String,
    onValueChange: (String) -> Unit,
    placeholder: String,
    leadingIcon: @Composable (() -> Unit)? = null,
    visualTransformation: VisualTransformation = VisualTransformation.None,
    isError: Boolean = false,
    supportingText: String? = null,
    keyboardOptions: KeyboardOptions = KeyboardOptions.Default
) {
    val leadingIconComposable: @Composable (() -> Unit)? = leadingIcon?.let {
        @Composable { it() }
    }

    val supportingTextComposable: @Composable (() -> Unit)? = supportingText?.let { text ->
        @Composable { Text(text = text) }
    }

    OutlinedTextField(
        modifier = modifier
            .fillMaxWidth(),
        value = value,
        onValueChange = onValueChange,
        placeholder = {
            Text(
                text = placeholder,
                color = TextSecondary
            )
        },
        singleLine = true,
        shape = RoundedCornerShape(16.dp),
        leadingIcon = leadingIconComposable,
        supportingText = supportingTextComposable,

        visualTransformation = visualTransformation,
        keyboardOptions = keyboardOptions,
        isError = isError,

        colors = OutlinedTextFieldDefaults.colors(
            focusedBorderColor = GreenPrimary,
            unfocusedBorderColor = BorderLight,
            focusedContainerColor = MaterialTheme.colorScheme.surface,
            unfocusedContainerColor = MaterialTheme.colorScheme.surface,
            cursorColor = GreenPrimary,
            focusedLeadingIconColor = GreenPrimary,
            unfocusedLeadingIconColor = Color.Gray
        )
    )
}